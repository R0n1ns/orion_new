package handlers

import (
	"fmt"
	"log"
	"net/http"
	"orion/data"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket Manager
type WS struct {
	Upgrader    websocket.Upgrader
	Connections map[uint]*websocket.Conn // Map of UserID to WebSocket connection
}

var WSmanager = WS{}

func init() {
	// Initialize WebSocket Upgrader
	WSmanager.Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins (adjust for production)
		},
	}

	// Initialize Connections Map
	WSmanager.Connections = make(map[uint]*websocket.Conn)
}

// Handle WebSocket connections
func (ws *WS) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract and validate JWT token
	userID, err := extractJWT(r)
	if err != nil {
		http.Error(w, "Unauthorized: Missing JWT", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP to WebSocket
	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	// Register connection
	ws.Connections[userID] = conn
	log.Printf("User %d connected via WebSocket\n", userID)

	// Handle incoming messages
	for {
		_, dt, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		typ, msg, _ := data.HandleRequest(dt, userID)
		switch typ {
		case "RcvdMessage":
			msg := msg.(data.RcvdMessage)
			mes := map[string]interface{}{
				"fromChatID": msg.ChatId,
				"message":    msg.Message,
				"timestamp":  time.Now(),
			}
			data.AddMessage(userID, msg.ChatId, msg.Message)
			ret := map[string]interface{}{"method": "RcvdMessage", "data": mes}
			fmt.Println(ws.Connections)
			fmt.Println(msg.ChatId)
			chatusers, er := data.GetUsersInChat(msg.ChatId)
			if er != nil {
				log.Println("Error getting users from database:", er)
			}
			fmt.Println(chatusers)
			for _, user := range chatusers {
				if user.ID == userID {
					continue
				}
				if conn, ok := ws.Connections[user.ID]; ok {
					err := conn.WriteJSON(ret)
					if err != nil {
						log.Printf("Error sending message to User %d: %v\n", msg.ChatId, err)
						conn.Close()
						delete(ws.Connections, msg.ChatId)
					} else {
						log.Printf("Message sent from User %d to User %d: %s\n", userID, msg.ChatId, msg.Message)
					}
				} else {
					log.Printf("No active WebSocket connection for User %d\n", msg.ChatId)
				}
			}

		case "GetChats":
			chats := msg.([]data.Channel)
			chatsJSON := make([]map[string]interface{}, len(chats))
			for i, chat := range chats {
				chatsJSON[i] = map[string]interface{}{
					"id":     chat.ID,
					"name":   chat.Name,
					"readed": data.IfReadedChat(chat.ID),
				}
			}

			ret := map[string]interface{}{"method": "GetChats", "data": chatsJSON}
			if conn, ok := ws.Connections[userID]; ok {
				err := conn.WriteJSON(ret)
				if err != nil {
					log.Printf("Error sending chats %d: %v\n", userID, err)
					conn.Close()
					delete(ws.Connections, userID)
				} else {
					log.Printf("Chats sendet to user %d \n", userID)
				}
			} else {
				log.Printf("No active WebSocket connection for User %d\n", userID)
			}
		case "GetChat":
			chat := msg.(data.GetChat)
			id := chat.ChatId
			allMasseges, err := data.GetChanMassages(uint(id))
			masseges := make([]map[string]string, 0, len(allMasseges))
			for _, massege := range allMasseges {
				masseges = append(masseges, map[string]string{
					"ID":         strconv.FormatUint(uint64(massege.ID), 10),
					"ChannelID":  strconv.FormatUint(uint64(massege.ChannelID), 10),
					"UserFromID": strconv.FormatUint(uint64(massege.UserID), 10),
					"Message":    massege.Content,
					"Edited":     strconv.FormatBool(massege.Edited),
					"Readed":     strconv.FormatBool(massege.Readed),
					"Date":       massege.Timestamp.Format("2 Jan 2006 15:04"),
				})
			}
			if err != nil {
				fmt.Println(err)
			}
			ret := map[string]interface{}{"method": "GetChat", "data": masseges}
			if conn, ok := ws.Connections[userID]; ok {
				err := conn.WriteJSON(ret)
				if err != nil {
					log.Printf("Error sending chats %d: %v\n", userID, err)
					conn.Close()
					delete(ws.Connections, userID)
				} else {
					log.Printf("Chats sendet to user %d \n", userID)
				}
			} else {
				log.Printf("No active WebSocket connection for User %d\n", userID)
			}
		}
		// Forward message to the intended recipient
	}

	// Cleanup on disconnect
	delete(ws.Connections, userID)
	log.Printf("User %d disconnected from WebSocket\n", userID)
}

// Forward message from sender to recipient
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
