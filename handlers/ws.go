package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"orion/data"
	"time"

	"github.com/gorilla/websocket"
)

// WS представляет менеджер WebSocket-соединений.
type WS struct {
	Upgrader    websocket.Upgrader
	Connections map[uint]*websocket.Conn // Соответствие между ID пользователя и его WebSocket-соединением.
}

// WSmanager – глобальный экземпляр менеджера WebSocket-соединений.
var WSmanager = WS{}

func SendCountConn() {
	ticker := time.NewTicker(time.Second * 5)
	for _ = range ticker.C {
		ActiveChatsGauge.Set(float64(len(WSmanager.Connections)))
	}
}

// init инициализирует менеджер WebSocket: устанавливает апгрейдер и инициализирует карту подключений.
func init() {
	WSmanager.Upgrader = websocket.Upgrader{
		// Разрешает подключение с любого источника; для продакшена рекомендуется ограничить список допустимых.
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	WSmanager.Connections = make(map[uint]*websocket.Conn)
	go SendCountConn()

}

// HandleWebSocket обрабатывает установку WebSocket-соединения и входящие сообщения от клиента.
// Функция извлекает JWT-токен для идентификации пользователя, обновляет карту соединений и
// выполняет обработку поступающих запросов.
func (ws *WS) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	ws.Connections[userID] = conn
	log.Printf("User %d connected", userID)

	defer func() {
		conn.Close()
		delete(ws.Connections, userID)
		log.Printf("User %d disconnected", userID)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		// Обработка сообщения
		var dt Querys
		if err := json.Unmarshal(message, &dt); err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}
		log.Println("Received request:", dt)

		startTime := time.Now()
		defer func(method string, start time.Time) {
			elapsed := time.Since(start)
			MessageProcessingTime.WithLabelValues(method).Observe(elapsed.Seconds())
		}(dt.Method, startTime)

		if dt.Method != "RcvdMessage" {
			log.Println("Unsupported method:", dt.Method)
			continue
		}

		dat, ok := dt.Query.(map[string]interface{})
		if !ok {
			log.Println("Invalid query format for RcvdMessage")
			continue
		}

		// Безопасное извлечение и проверка chatId
		var chatId uint
		if chatRaw, ok := dat["chatId"]; ok {
			if chatFloat, ok := chatRaw.(float64); ok && chatFloat > 0 {
				chatId = uint(chatFloat)
			}
		}

		messageText, _ := dat["message"].(string)
		msg := RcvdMessage{ChatId: chatId, Message: messageText}

		// Если chatId не передан — создаём новый чат
		if chatId == 0 {
			user2Raw, ok := dat["user2"]
			if !ok {
				log.Println("user2 not provided for new chat creation")
				continue
			}
			user2ID := uint(user2Raw.(float64))

			newChat, err := data.CreateChat(userID, user2ID, "еые")
			if err != nil {
				log.Println("CreateChat error:", err)
				continue
			}
			chatId = newChat.ID
			msg.ChatId = chatId
		}

		// Добавляем сообщение
		data.AddMessage(userID, chatId, msg.Message)

		// Формируем ответ
		resp := map[string]interface{}{
			"method": "RcvdMessage",
			"data": map[string]interface{}{
				"fromChatID": chatId,
				"UserFromID": userID,
				"message":    msg.Message,
				"timestamp":  time.Now().Format(time.RFC3339),
				"Readed":     false,
			},
		}

		users, err := data.GetUsersInChat(chatId)
		if err != nil {
			log.Println("GetUsersInChat error:", err)
			continue
		}

		for _, user := range users {
			if c, ok := ws.Connections[user.ID]; ok {
				if err := c.WriteJSON(resp); err != nil {
					log.Printf("Failed to send to user %d: %v", user.ID, err)
					c.Close()
					delete(ws.Connections, user.ID)
				}
			}
		}
	}
}
