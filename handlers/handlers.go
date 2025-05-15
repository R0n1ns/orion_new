package handlers

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"orion/data/manager"
	"orion/services/jwt"
	"orion/services/minio"
	"strconv"
	"strings"
	"time"
)

// GetChatsHandler возвращает список чатов и информацию о пользователе.
func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chats, err := manager.GetChannels(userID)
	if err != nil {
		http.Error(w, "failed to fetch chats", http.StatusInternalServerError)
		return
	}

	chatsJSON := make([]map[string]interface{}, 0)

	for _, chat := range chats {
		users, _ := manager.GetUsersInChat(chat.ID)

		// Формируем список пользователей
		userList := make([]map[string]interface{}, 0)
		var chatName string
		var profilePicture string
		for _, user := range users {
			userList = append(userList, map[string]interface{}{
				"id":       user.ID,
				"username": user.UserName,
			})

			if user.ID != userID {
				chatName = user.UserName
				profilePicture = minio.GetPhoto(user.ProfilePicture)
			}
		}

		chatJSON := map[string]interface{}{
			"id":              chat.ID,
			"name":            chatName,
			"readed":          manager.IfReadedChat(chat.ID, userID),
			"Private":         chat.IsPrivate,
			"profile_picture": profilePicture,
			"users":           userList,
		}

		chatsJSON = append(chatsJSON, chatJSON)
	}

	user := manager.GetUserByID(userID)
	resp := map[string]interface{}{
		"chats": chatsJSON,
		"info": map[string]interface{}{
			"id":             user.ID,
			"Mail":           user.Mail,
			"UserName":       user.UserName,
			"IsBlocked":      user.IsBlocked,
			"LastOnline":     user.LastOnline,
			"ProfilePicture": minio.GetPhoto(user.ProfilePicture),
			"Biom":           user.Bio,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// GetUsersHandler возвращает пользователей по подстроке имени.
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := r.URL.Query().Get("username")
	users := manager.SearchByUsername(username)
	res := []map[string]interface{}{}
	for _, u := range users {
		if u.ID == userID {
			continue // не показываем себя
		}
		res = append(res, map[string]interface{}{
			"id":              u.ID,
			"username":        u.UserName,
			"profile_picture": minio.GetPhoto(u.ProfilePicture), // Полный URL
			"chat_id":         manager.GetChatIDForUsers(userID, u.ID),
		})
	}
	json.NewEncoder(w).Encode(res)
}

// CreateChatHandler создаёт чат между двумя пользователями.
func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
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
	chat, err := manager.CreateChat(userID, b.User2, chatName)
	if err != nil {
		http.Error(w, "cannot create chat", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(chat)
}

// GetChatMessagesHandler возвращает все сообщения в чате.
func GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	//userID, err := jwt.ExtractJWT(w, r)
	//if err != nil {
	//	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	//	return
	//}
	chatIDStr := r.URL.Query().Get("chatId")
	chatID, err := strconv.Atoi(chatIDStr)

	if err != nil || chatID == -1 {
		messagesJSON := []map[string]interface{}{}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"chatId":   chatID,
			"messages": messagesJSON,
		})
	} else {

		msgs, err := manager.GetChanMassages(uint(chatID))
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

}

// MarkMessagesReadHandler помечает сообщения как прочитанные.
func MarkMessagesReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
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
	manager.ReadMessages(float64(chatID), strconv.Itoa(int(userID)))
	w.WriteHeader(http.StatusNoContent)
}

// UpdateProfileHandler обновляет профиль пользователя.
func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
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
	err = manager.UpdateUser(userID, b.Mail, b.UserName, b.Biom)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UploadProfilePictureHandler обрабатывает загрузку фото профиля.
func UploadProfilePictureHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
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
	manager.AddHexPhoto(userID, imageHex)
	objectName := imageHex + ".jpg"
	err = minio.UploadImage(context.Background(), objectName, imageBytes, "image/jpeg")
	if err != nil {
		http.Error(w, "upload error", http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"ProfilePicture": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageBytes),
	}
	json.NewEncoder(w).Encode(resp)
}

func BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	type Request struct {
		UserID string `json:"userId"` // строка
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	parsedID, err := strconv.ParseUint(req.UserID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	//log.Println(json.NewDecoder(r.Body))

	if err := manager.BlockUser(userID, uint(parsedID)); err != nil {
		http.Error(w, "Block failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	type Request struct {
		UserID string `json:"userId"` // строка
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	parsedID, err := strconv.ParseUint(req.UserID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	if err := manager.UnblockUser(userID, uint(parsedID)); err != nil {
		http.Error(w, "Unblock failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CheckUserBlockedHandler возвращает статус блокировки пользователя (поле IsBlocked)
func CheckUserBlockedHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID целевого пользователя из query-параметра
	targetUserIDStr := r.URL.Query().Get("userId")
	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Получаем информацию о пользователе
	targetUser := manager.GetUserByID(uint(targetUserID))
	if targetUser.ID == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Возвращаем статус блокировки
	json.NewEncoder(w).Encode(map[string]bool{
		"isBlocked": targetUser.IsBlocked,
	})
}

// CheckMutualBlockHandler проверяет, заблокированы ли пользователи друг другом
func CheckMutualBlockHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Получаем chatId из запроса
	chatIDStr := r.URL.Query().Get("chatId")
	log.Println("chatIDStr", chatIDStr)
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil || chatID <= 0 {
		http.Error(w, "Invalid chatId", http.StatusBadRequest)
		return
	}

	// 2. Получаем пользователей в чате
	users, err := manager.GetUsersInChat(uint(chatID))
	if err != nil || len(users) != 2 {
		http.Error(w, "Chat not found or invalid", http.StatusBadRequest)
		return
	}

	// 3. Определяем ID другого пользователя
	var targetUserID uint
	for _, user := range users {
		if user.ID != userID {
			targetUserID = user.ID
			break
		}
	}

	if targetUserID == 0 {
		http.Error(w, "User not found in chat", http.StatusBadRequest)
		return
	}

	// 4. Проверяем взаимную блокировку
	isBlockedByCurrent := manager.IsBlocked(userID, targetUserID)
	isBlockedByTarget := manager.IsBlocked(targetUserID, userID)

	json.NewEncoder(w).Encode(map[string]bool{
		"isMutuallyBlocked": isBlockedByCurrent || isBlockedByTarget,
	})
}
