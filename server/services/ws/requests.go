package ws

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

// RcvdMessage описывает структуру запроса на отправку сообщения/изображения
type RcvdMessage struct {
	ChatId  int    `json:"chatId"`
	Message string `json:"message,omitempty"` // Текст (опционально)
	Image   string `json:"image,omitempty"`   // Base64 (опционально)
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
