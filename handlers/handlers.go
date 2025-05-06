package handlers

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"orion/data"
	"strconv"
	"strings"
	"time"
)

// GetChatsHandler возвращает список чатов и информацию о пользователе.
func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chats, err := data.GetChannels(userID)
	if err != nil {
		http.Error(w, "failed to fetch chats", http.StatusInternalServerError)
		return
	}

	chatsJSON := make([]map[string]interface{}, len(chats))
	for i, chat := range chats {
		users, _ := data.GetUsersInChat(chat.ID)
		for _, user := range users {
			if user.ID != userID {
				chatsJSON[i] = map[string]interface{}{
					"id":              chat.ID,
					"name":            user.UserName,
					"readed":          data.IfReadedChat(chat.ID, userID),
					"Private":         chat.IsPrivate,
					"profile_picture": data.GetPhoto(user.ProfilePicture),
				}
			}
		}
	}

	user := data.GetUserByID(userID)
	resp := map[string]interface{}{
		"chats": chatsJSON,
		"info": map[string]interface{}{
			"id":             user.ID,
			"Mail":           user.Mail,
			"UserName":       user.UserName,
			"IsBlocked":      user.IsBlocked,
			"LastOnline":     user.LastOnline,
			"ProfilePicture": data.GetPhoto(user.ProfilePicture),
			"Biom":           user.Bio,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// GetUsersHandler возвращает пользователей по подстроке имени.
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := r.URL.Query().Get("username")
	users := data.SearchByUsername(username)
	res := []map[string]interface{}{}
	for _, u := range users {
		if u.ID == userID {
			continue // не показываем себя
		}
		res = append(res, map[string]interface{}{
			"id":              u.ID,
			"username":        u.UserName,
			"profile_picture": data.GetPhoto(u.ProfilePicture), // Полный URL
			"chat_id":         data.GetChatIDForUsers(userID, u.ID),
		})
	}
	json.NewEncoder(w).Encode(res)
}

// CreateChatHandler создаёт чат между двумя пользователями.
func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	type body struct {
		User2 uint `json:"user2"`
	}
	var b body
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	chatName := strconv.Itoa(int(userID)) + "--" + strconv.Itoa(int(b.User2))
	chat, err := data.CreateChat(userID, b.User2, chatName)
	if err != nil {
		http.Error(w, "cannot create chat", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(chat)
}

// GetChatMessagesHandler возвращает все сообщения в чате.
func GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	chatIDStr := r.URL.Query().Get("chatId")
	chatID, err := strconv.Atoi(chatIDStr)

	if err != nil || chatID == -1 {

		// Получаем ID второго участника из запроса
		otherIDStr := r.URL.Query().Get("otherUserId")
		otherID, _ := strconv.Atoi(otherIDStr)
		log.Println("userID", userID)
		log.Println("otherID", otherID)

		newChat, err := data.CreateChat(userID, uint(otherID), "test")
		if err != nil {
			http.Error(w, "не удалось создать чат", 500)
			return
		}

		chatID = int(newChat.ID)
	}

	msgs, err := data.GetChanMassages(uint(chatID))
	if err != nil {
		http.Error(w, "cannot get messages", http.StatusInternalServerError)
		return
	}

	messagesJSON := []map[string]interface{}{}
	for _, m := range msgs {
		messagesJSON = append(messagesJSON, map[string]interface{}{
			"id":        m.ID,
			"from":      m.UserID,
			"message":   m.Content,
			"timestamp": m.Timestamp.Format(time.RFC3339),
			"readed":    m.Readed,
		})
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chatId":   chatID,
		"messages": messagesJSON,
	})
}

// MarkMessagesReadHandler помечает сообщения как прочитанные.
func MarkMessagesReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	chatIDStr := r.URL.Query().Get("chatId")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		http.Error(w, "invalid chatId", http.StatusBadRequest)
		return
	}
	data.ReadMessages(float64(chatID), strconv.Itoa(int(userID)))
	w.WriteHeader(http.StatusNoContent)
}

// UpdateProfileHandler обновляет профиль пользователя.
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	type body struct {
		Mail     string `json:"Mail"`
		UserName string `json:"UserName"`
		Biom     string `json:"Biom"`
	}
	var b body
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	err = data.UpdateUser(userID, b.Mail, b.UserName, b.Biom)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UploadProfilePictureHandler обрабатывает загрузку фото профиля.
func UploadProfilePictureHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	type body struct {
		ImageData string `json:"imageData"`
	}
	var b body
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	parts := strings.SplitN(b.ImageData, ",", 2)
	if len(parts) != 2 {
		http.Error(w, "invalid image data", http.StatusBadRequest)
		return
	}
	imageBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		http.Error(w, "base64 decode error", http.StatusBadRequest)
		return
	}
	hash := md5.Sum(imageBytes)
	imageHex := hex.EncodeToString(hash[:])
	//log.Println(userID)
	//log.Println(imageHex)
	data.AddHexPhoto(userID, imageHex)
	objectName := imageHex + ".jpg"
	err = data.UploadImage(context.Background(), objectName, imageBytes, "image/jpeg")
	if err != nil {
		http.Error(w, "upload error", http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"ProfilePicture": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageBytes),
	}
	json.NewEncoder(w).Encode(resp)
}
