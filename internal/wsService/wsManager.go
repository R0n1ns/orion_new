package wsService

import (
	"fmt"
)

type WebSocketManager struct {
	handler RequestHandler
}

func NewWebSocketManager() *WebSocketManager {
	validator := &RcvdMessageValidator{} // Используем указатель
	handler := RequestHandler{Validator: validator}
	return &WebSocketManager{handler: handler}
}

func (wm *WebSocketManager) HandleMessage(data []byte, userID uint) {
	method, response, err := wm.handler.HandleRequest(data, userID)
	if err != nil {
		fmt.Println("Error processing message:", err)
		// Отправка ошибки клиенту
		return
	}

	// Отправить ответ клиенту
	fmt.Println("Method:", method, "Response:", response)
}
