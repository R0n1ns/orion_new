package handlers

import (
	"log"
	"net/http"
	"orion/data"
	"strconv"
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
	// Извлечение и проверка JWT-токена (функция extractJWT должна быть реализована отдельно).
	userID, err := extractJWT(w, r)
	if err != nil {
		http.Error(w, "Unauthorized: Missing or invalid JWT", http.StatusUnauthorized)
		return
	}

	// Обновление соединения: перевод HTTP в WebSocket.
	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	// Регистрируем соединение.
	ws.Connections[userID] = conn
	log.Printf("User %d connected via WebSocket\n", userID)

	// Цикл чтения входящих сообщений.
	for {
		_, dt, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Обработка запроса.
		respType, msg, err := HandleRequest(dt, userID)
		if err != nil {
			log.Println("HandleRequest error:", err)
			continue
		}

		switch respType {
		case "RcvdMessage":
			// Обработка отправки сообщения.
			receivedMsg := msg.(RcvdMessage)
			// Сохранение сообщения в БД.
			data.AddMessage(userID, receivedMsg.ChatId, receivedMsg.Message)
			// Формирование ответа для отправки в чат.
			response := map[string]interface{}{
				"method": "RcvdMessage",
				"data": map[string]interface{}{
					"fromChatID": receivedMsg.ChatId,
					"message":    receivedMsg.Message,
					"timestamp":  time.Now(),
				},
			}
			// Получение списка пользователей, участвующих в чате.
			chatUsers, err := data.GetUsersInChat(receivedMsg.ChatId)
			if err != nil {
				log.Println("Error getting users from database:", err)
				break
			}
			// Отправка сообщения всем участникам, кроме отправителя.
			for _, user := range chatUsers {
				if user.ID == userID {
					continue
				}
				if conn, ok := ws.Connections[user.ID]; ok {
					if err := conn.WriteJSON(response); err != nil {
						log.Printf("Error sending message to User %d: %v\n", user.ID, err)
						conn.Close()
						delete(ws.Connections, user.ID)
					} else {
						log.Printf("Message sent from User %d to User %d: %s\n", userID, user.ID, receivedMsg.Message)
					}
				} else {
					log.Printf("No active WebSocket connection for User %d\n", user.ID)
				}
			}

		case "ReadedMessages":
			// Отправка уведомления о прочтении сообщений.
			chatID := uint(msg.(float64))
			response := map[string]interface{}{
				"method": "ReadedMessages",
				"data":   map[string]interface{}{"chatId": chatID},
			}
			chatUsers, err := data.GetUsersInChat(chatID)
			if err != nil {
				log.Println("Error getting users from database:", err)
				break
			}
			for _, user := range chatUsers {
				if user.ID == userID {
					continue
				}
				if conn, ok := ws.Connections[user.ID]; ok {
					if err := conn.WriteJSON(response); err != nil {
						log.Printf("Error sending read notification to User %d: %v\n", user.ID, err)
						conn.Close()
						delete(ws.Connections, user.ID)
					} else {
						log.Printf("Read notification sent from User %d to User %d\n", userID, user.ID)
					}
				} else {
					log.Printf("No active WebSocket connection for User %d\n", user.ID)
				}
			}

		case "GetChats":
			// Формирование ответа со списком чатов.
			chats := msg.([]data.Channel)
			chatsJSON := make([]map[string]interface{}, len(chats))
			for i, chat := range chats {
				chatUsers, _ := data.GetUsersInChat(chat.ID)
				if len(chatUsers) == 1 {
					chatsJSON[i] = map[string]interface{}{
						"id":      chat.ID,
						"name":    chatUsers[0].UserName,
						"readed":  data.IfReadedChat(chat.ID, userID),
						"Private": chat.IsPrivate,
					}
				} else {
					// Для личных чатов берётся имя второго участника.
					for _, user := range chatUsers {
						if user.ID == userID {
							continue
						}
						chatsJSON[i] = map[string]interface{}{
							"id":      chat.ID,
							"name":    user.UserName,
							"readed":  data.IfReadedChat(chat.ID, userID),
							"Private": chat.IsPrivate,
						}
					}
				}
			}
			userC := data.GetUserByID(userID)
			response := map[string]interface{}{
				"method": "GetChats",
				"data": map[string]interface{}{
					"chats": chatsJSON,
					"info": map[string]interface{}{
						"id":             userC.ID,
						"Mail":           userC.Mail,
						"UserName":       userC.UserName,
						"IsBlocked":      userC.IsBlocked,
						"LastOnline":     userC.LastOnline,
						"ProfilePicture": data.GetPhoto(userC.ProfilePicture),
						"Biom":           userC.Bio,
					},
				},
			}
			if conn, ok := ws.Connections[userID]; ok {
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("Error sending chats to User %d: %v\n", userID, err)
					conn.Close()
					delete(ws.Connections, userID)
				} else {
					log.Printf("Chats sent to user %d\n", userID)
				}
			} else {
				log.Printf("No active WebSocket connection for User %d\n", userID)
			}

		case "UpdateProfilePicture":
			// Отправка обновлённого URL фотографии профиля.
			profilePicURL := msg.(string)
			response := map[string]interface{}{
				"method": "UpdateProfilePicture",
				"data":   map[string]interface{}{"ProfilePicture": profilePicURL},
			}
			if conn, ok := ws.Connections[userID]; ok {
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("Error sending profile picture update to User %d: %v\n", userID, err)
					conn.Close()
					delete(ws.Connections, userID)
				} else {
					log.Printf("Profile picture update sent to user %d\n", userID)
				}
			} else {
				log.Printf("No active WebSocket connection for User %d\n", userID)
			}

		case "ChatCreated":
			// Уведомление об успешном создании чата.
			chat := msg.(data.Channel)
			var chatResponse map[string]interface{}
			for _, user := range chat.Users {
				if user.ID == userID {
					continue
				}
				chatResponse = map[string]interface{}{
					"id":      chat.ID,
					"name":    user.UserName,
					"readed":  data.IfReadedChat(chat.ID, userID),
					"Private": chat.IsPrivate,
				}
			}
			response := map[string]interface{}{
				"method": "ChatCreated",
				"data":   chatResponse,
			}
			if conn, ok := ws.Connections[userID]; ok {
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("Error sending chat creation notification to User %d: %v\n", userID, err)
					conn.Close()
					delete(ws.Connections, userID)
				} else {
					log.Printf("Chat creation notification sent to user %d\n", userID)
				}
			} else {
				log.Printf("No active WebSocket connection for User %d\n", userID)
			}

		case "GetUsers":
			// Обработка ответа на запрос пользователей.
			users := msg.([]data.User)
			usersJSON := make([]map[string]interface{}, len(users))
			for i, user := range users {
				usersJSON[i] = map[string]interface{}{
					"id":       user.ID,
					"username": user.UserName,
					"chat_id":  data.GetChatIDForUsers(user.ID, userID),
				}
			}
			response := map[string]interface{}{
				"method": "UsersAnsw",
				"data":   usersJSON,
			}
			if conn, ok := ws.Connections[userID]; ok {
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("Error sending users answer to User %d: %v\n", userID, err)
					conn.Close()
					delete(ws.Connections, userID)
				} else {
					log.Printf("Users answer sent to user %d\n", userID)
				}
			} else {
				log.Printf("No active WebSocket connection for User %d\n", userID)
			}

		case "GetChat":
			// Формирование ответа на запрос данных конкретного чата.
			chatReq := msg.(GetChat)
			chatID := chatReq.ChatId
			allMessages, err := data.GetChanMassages(chatID)
			if err != nil {
				log.Println("Error getting messages:", err)
				break
			}
			// Преобразование сообщений в формат для отправки.
			messagesJSON := make([]map[string]interface{}, 0, len(allMessages))
			var lastReadedID uint
			lastReadedFound := false
			for _, m := range allMessages {
				// Если предыдущее сообщение не прочитано, запоминаем его ID.
				if lastReadedFound && !m.Readed {
					lastReadedID = m.ID
				}
				lastReadedFound = m.Readed
				messagesJSON = append(messagesJSON, map[string]interface{}{
					"ID":         strconv.FormatUint(uint64(m.ID), 10),
					"ChannelID":  strconv.FormatUint(uint64(m.ChannelID), 10),
					"UserFromID": strconv.FormatUint(uint64(m.UserID), 10),
					"Message":    m.Content,
					"Edited":     strconv.FormatBool(m.Edited),
					"Readed":     strconv.FormatBool(m.Readed),
					"Date":       m.Timestamp.Format("2 Jan 2006 15:04"),
				})
			}
			chatData := data.GetChatByID(chatID)
			// Если чат является приватным, определяем статус онлайн другого участника.
			var online bool
			var otherUser data.User
			if chatData.IsPrivate {
				users, _ := data.GetUsersInChat(chatData.ID)
				for _, user := range users {
					if user.ID != userID {
						otherUser = user
						_, online = ws.Connections[user.ID]
					}
				}
			}
			response := map[string]interface{}{
				"method": "GetChat",
				"data": map[string]interface{}{
					"messages":     messagesJSON,
					"Online":       online,
					"LastReadedId": lastReadedID,
					"info": map[string]interface{}{
						"Mail":           otherUser.Mail,
						"UserName":       otherUser.UserName,
						"IsBlocked":      otherUser.IsBlocked,
						"LastOnline":     otherUser.LastOnline,
						"ProfilePicture": data.GetPhoto(otherUser.ProfilePicture),
						"Biom":           otherUser.Bio,
					},
				},
			}
			if conn, ok := ws.Connections[userID]; ok {
				if err := conn.WriteJSON(response); err != nil {
					log.Printf("Error sending chat data to User %d: %v\n", userID, err)
					conn.Close()
					delete(ws.Connections, userID)
				} else {
					log.Printf("Chat data sent to user %d\n", userID)
				}
			} else {
				log.Printf("No active WebSocket connection for User %d\n", userID)
			}

		default:
			log.Printf("Unknown response type: %s\n", respType)
		}
	}

	// Очистка соединения при разрыве.
	delete(ws.Connections, userID)
	log.Printf("User %d disconnected from WebSocket\n", userID)
}

// forwardMessage пересылает сообщение от одного пользователя к другому через WebSocket.
// Если у получателя нет активного соединения, выводится сообщение в лог.
func (ws *WS) forwardMessage(fromUserID, toUserID uint, message string) {
	if conn, ok := ws.Connections[toUserID]; ok {
		err := conn.WriteJSON(map[string]interface{}{
			"fromChatID": fromUserID,
			"message":    message,
			"timestamp":  time.Now(),
		})
		if err != nil {
			log.Printf("Error sending message to User %d: %v\n", toUserID, err)
			conn.Close()
			delete(ws.Connections, toUserID)
		} else {
			log.Printf("Message sent from User %d to User %d: %s\n", fromUserID, toUserID, message)
		}
	} else {
		log.Printf("No active WebSocket connection for User %d\n", toUserID)
	}
}
