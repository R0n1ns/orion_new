package wsService

import (
	"errors"
)

// Querys представляет общий формат запроса от клиента.
type Querys struct {
	Method string      `json:"method"`
	Query  interface{} `json:"query"`
}

// Валидатор запроса
type RequestValidator interface {
	Validate(data map[string]interface{}) error
}

type RcvdMessageValidator struct{}

func (v *RcvdMessageValidator) Validate(data map[string]interface{}) error {
	if _, ok := data["message"].(string); !ok {
		return errors.New("message is required and must be a string")
	}
	if _, ok := data["chatId"].(float64); !ok {
		return errors.New("chatId is required and must be a number")
	}
	return nil
}

// Структуры для различных типов запросов
type RcvdMessage struct {
	ChatId  uint   `json:"chatId"`
	Message string `json:"message"`
}

type GetChat struct {
	ChatId uint `json:"chatId"`
}

type Default struct {
	Nothing string `json:"nothing"`
}
