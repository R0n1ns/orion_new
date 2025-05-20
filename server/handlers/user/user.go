package user

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"orion/server/data/manager"
	"orion/server/services/jwt"
	"orion/server/services/minio"
	"orion/server/services/ws"
	"strconv"
	"strings"
)

// RegisterRoutes регистрирует маршруты чата на переданном роутере
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/profile", UpdateProfileHandler).Methods("PUT")
	r.HandleFunc("/api/profile/photo", UploadProfilePictureHandler).Methods("POST")
	r.HandleFunc("/api/block", BlockUserHandler).Methods("POST")
	r.HandleFunc("/api/unblock", UnblockUserHandler).Methods("POST")
	r.HandleFunc("/api/block-status", CheckUserBlockedHandler).Methods("GET")
	r.HandleFunc("/api/mutual-block", CheckMutualBlockHandler).Methods("GET")
	r.HandleFunc("/api/online-status", OnlineStatusHandler).Methods("GET")
	r.HandleFunc("/api/users", GetUsersHandler).Methods("GET")
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

// OnlineStatusHandler возвращает статус пользователя
func OnlineStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := jwt.ExtractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	//targetUserIDStr := r.URL.Query().Get("userId")
	//targetUserID, _ := strconv.Atoi(targetUserIDStr)

	isOnline := ws.WSmanager.Connections[userID] != nil
	user := manager.GetUserByID(userID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"isOnline":   isOnline,
		"lastOnline": user.LastOnline,
	})
}
