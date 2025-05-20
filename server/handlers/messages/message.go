package messages

import (
	"github.com/gorilla/mux"
	"net/http"
	"orion/data/manager"
	"orion/server/services/jwt"
	"strconv"
)

// RegisterRoutes регистрирует маршруты чата на переданном роутере
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/messages/read", MarkMessagesReadHandler).Methods("POST", "PUT")
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
