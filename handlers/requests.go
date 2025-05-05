package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Querys представляет общий формат запроса от клиента.
//
// Пример запроса:
//
//	{
//	  "method": "RcvdMessage",
//	  "query": { "chatId": 1, "message": "Привет" }
//	}
type Querys struct {
	Method string      `json:"method"`
	Query  interface{} `json:"query"`
}

// RcvdMessage описывает структуру запроса на отправку сообщения от клиента.
//
// Пример:
//
//	{ "method": "RcvdMessage", "query": { "chatId": 1, "message": "Текст сообщения" } }
type RcvdMessage struct {
	ChatId  uint   `json:"chatId"`
	Message string `json:"message"`
}

// GetChat описывает запрос для получения информации о конкретном чате.
//
// Пример:
//
//	{ "method": "GetChat", "query": { "chatId": 1 } }
type GetChat struct {
	ChatId uint `json:"chatId"`
}

// Default представляет структуру по умолчанию (при отсутствии данных).
type Default struct {
	Nothing string `json:"nothing"`
}

// HandleRequest обрабатывает входящий JSON-запрос от клиента и выполняет соответствующие действия.
// При этом функция возвращает:
//   - тип ответа (метод, который затем используется для формирования ответа клиенту),
//   - данные ответа,
//   - ошибку (если произошла).
//
// Параметры:
//   - data: входящие данные в формате JSON.
//   - userID: ID пользователя, с которым ассоциирован запрос (например, из JWT).
//
// Возможные методы запроса:
//   - "RcvdMessage" – получение сообщения для отправки;
//   - "ReadMessages" – отметить сообщения как прочитанные;
//   - "GetChats" – получить список чатов пользователя;
//   - "UpdateProfilePicture" – обновить фотографию профиля;
//   - "GetUsers" – поиск пользователей по имени;
//   - "CreateChat" – создание нового чата;
//   - "GetChat" – получить данные конкретного чата.
func HandleRequest(data []byte, userID uint) (string, interface{}, map[string]interface{}, error) {
	var dt Querys
	if err := json.Unmarshal(data, &dt); err != nil {
		return "", nil, map[string]interface{}{}, fmt.Errorf("failed to unmarshal request: %w", err)
	}
	fmt.Println("Received request:", dt)

	//отправк метрики
	startTime := time.Now()
	defer func(method string, start time.Time) {
		elapsed := time.Since(start)
		// Отправляем время обработки в секундах с меткой, указывающей тип сообщения.
		MessageProcessingTime.WithLabelValues(method).Observe(elapsed.Seconds())
	}(dt.Method, startTime)

	switch dt.Method {
	case "RcvdMessage":
		var msg RcvdMessage
		dat, ok := dt.Query.(map[string]interface{})
		if !ok {
			return "", nil, map[string]interface{}{}, errors.New("invalid query format for RcvdMessage")
		}
		msg.Message = dat["message"].(string)
		msg.ChatId = uint(dat["chatId"].(float64))
		return "RcvdMessage", msg, dat, nil
	default:
		return "", nil, map[string]interface{}{}, fmt.Errorf("unsupported method for WebSocket: %s", dt.Method)
	}

}
