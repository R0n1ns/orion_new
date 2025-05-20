package chat

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"orion/server/data/manager"
	"orion/server/services/jwt"
	"orion/server/services/minio"
	"orion/server/services/ws"
	"strconv"
	"time"
)

// RegisterRoutes регистрирует маршруты чата на переданном роутере
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/chats", GetChatsHandler).Methods("GET")
	r.HandleFunc("/api/chat", CreateChatHandler).Methods("POST")
	r.HandleFunc("/api/messages", GetChatMessagesHandler).Methods("GET")
}

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
	user := manager.GetUserByID(userID)

	for _, chat := range chats {
		users, _ := manager.GetUsersInChat(chat.ID)
		var chatName string
		var profilePicture string
		var otherUserID uint
		var lastOnline time.Time
		var isOnline bool

		// Формируем расширенную информацию о пользователях
		userList := make([]map[string]interface{}, 0)
		for _, user := range users {
			userData := map[string]interface{}{
				"id":          user.ID,
				"username":    user.UserName,
				"is_online":   ws.WSmanager.Connections[user.ID] != nil,
				"last_online": user.LastOnline.Format(time.RFC3339),
			}

			if user.ID != userID {
				chatName = user.UserName
				profilePicture = minio.GetPhoto(user.ProfilePicture)
				otherUserID = user.ID
				lastOnline = user.LastOnline
				isOnline = ws.WSmanager.Connections[user.ID] != nil
			}

			userList = append(userList, userData)
		}

		chatJSON := map[string]interface{}{
			"id":              chat.ID,
			"name":            chatName,
			"readed":          manager.IfReadedChat(chat.ID, userID),
			"is_private":      chat.IsPrivate,
			"profile_picture": profilePicture,
			"users":           userList,
			"other_user_id":   otherUserID,
			"last_activity":   lastOnline.Format(time.RFC3339),
			"is_online":       isOnline,
			"unread_count":    manager.GetUnreadCount(chat.ID, userID),
			"Bio":             user.Bio,
		}

		chatsJSON = append(chatsJSON, chatJSON)
	}

	resp := map[string]interface{}{
		"chats": chatsJSON,
		"info": map[string]interface{}{
			"id":             user.ID,
			"Mail":           user.Mail,
			"UserName":       user.UserName,
			"IsBlocked":      user.IsBlocked,
			"LastOnline":     user.LastOnline.Format(time.RFC3339),
			"ProfilePicture": minio.GetPhoto(user.ProfilePicture),
			"Bio":            user.Bio,
		},
	}

	json.NewEncoder(w).Encode(resp)
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
