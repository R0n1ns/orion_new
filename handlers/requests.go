package handlers

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	data2 "orion/data"
	"strconv"
	"strings"
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
func HandleRequest(data []byte, userID uint) (string, interface{}, error) {
	var dt Querys
	if err := json.Unmarshal(data, &dt); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}
	fmt.Println("Received request:", dt)

	switch dt.Method {
	case "RcvdMessage":
		// Обработка запроса на отправку сообщения.
		var msg RcvdMessage
		dat, ok := dt.Query.(map[string]interface{})
		if !ok {
			return "", nil, errors.New("invalid query format for RcvdMessage")
		}
		msg.Message = dat["message"].(string)
		msg.ChatId = uint(dat["chatId"].(float64))
		// Здесь можно добавить вызов функции для сохранения сообщения в БД
		return "RcvdMessage", msg, nil

	case "ReadMessages":
		// Обработка запроса на пометку сообщений как прочитанных.
		dat, ok := dt.Query.(map[string]interface{})
		if !ok {
			return "", nil, errors.New("invalid query format for ReadMessages")
		}
		chatID, ok := dat["chatId"].(float64)
		if !ok {
			return "", nil, errors.New("chatId must be a number")
		}
		userIdStr, ok := dat["userId"].(string)
		if !ok {
			return "", nil, errors.New("userId must be a string")
		}
		data2.ReadMessages(chatID, userIdStr)
		return "ReadedMessages", chatID, nil

	case "GetChats":
		// Обработка запроса на получение списка чатов для пользователя.
		chats, err := data2.GetChannels(userID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get channels: %w", err)
		}
		return "GetChats", chats, nil

	case "UpdateProfilePicture":
		// Обработка запроса на обновление фотографии профиля.
		dataMap, ok := dt.Query.(map[string]interface{})
		if !ok {
			return "", nil, errors.New("invalid query format for UpdateProfilePicture")
		}
		imageData, ok := dataMap["imageData"].(string)
		if !ok {
			return "", nil, errors.New("imageData not provided or invalid")
		}
		parts := strings.SplitN(imageData, ",", 2)
		if len(parts) != 2 {
			return "", nil, errors.New("invalid image data format")
		}

		// Декодирование base64-части изображения.
		imageBytes, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return "", nil, fmt.Errorf("base64 decoding error: %w", err)
		}
		// Вычисляем MD5-хэш для имени файла.
		hash := md5.Sum(imageBytes)
		imageHex := hex.EncodeToString(hash[:])
		// Обновление ссылки на фотографию в базе данных.
		data2.AddHexPhoto(userID, imageHex)

		// Загрузка изображения в MinIO (заменяет сохранение на диск).
		objectName := imageHex + ".jpg"
		err = data2.UploadImage(context.Background(), objectName, imageBytes, "image/jpeg")
		if err != nil {
			return "", nil, fmt.Errorf("failed to upload image to minio: %w", err)
		}

		// Формирование data URL для изображения.
		encodedImage := base64.StdEncoding.EncodeToString(imageBytes)
		profilePictureURL := "data:image/jpeg;base64," + encodedImage
		return "UpdateProfilePicture", profilePictureURL, nil

	case "GetUsers":
		// Обработка запроса на поиск пользователей по имени.
		dat, ok := dt.Query.(map[string]interface{})
		if !ok {
			return "", nil, errors.New("invalid query format for GetUsers")
		}
		username, ok := dat["username"].(string)
		if !ok {
			return "", nil, errors.New("username must be provided as string")
		}
		users := data2.SearchByUsername(username)
		return "GetUsers", users, nil

	case "CreateChat":
		// Обработка запроса на создание нового чата.
		dat, ok := dt.Query.(map[string]interface{})
		if !ok {
			return "", nil, errors.New("invalid query format for CreateChat")
		}
		user1Str, ok := dat["user1"].(string)
		if !ok {
			return "", nil, errors.New("user1 must be a string")
		}
		user2Float, ok := dat["user2"].(float64)
		if !ok {
			return "", nil, errors.New("user2 must be a number")
		}
		chatname := user1Str + "--" + fmt.Sprint(user2Float)
		us1, err := strconv.ParseUint(user1Str, 10, 32)
		if err != nil {
			return "", nil, fmt.Errorf("failed to parse user1: %w", err)
		}
		chat, err := data2.CreateChat(uint(us1), uint(user2Float), chatname)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create chat: %w", err)
		}
		return "ChatCreated", *chat, nil

	case "GetChat":
		// Обработка запроса на получение данных конкретного чата.
		var getChatReq GetChat
		dat, ok := dt.Query.(map[string]interface{})
		if !ok {
			return "", nil, errors.New("invalid query format for GetChat")
		}
		chatID, ok := dat["chatId"].(float64)
		if !ok {
			return "", nil, errors.New("chatId must be a number")
		}
		getChatReq.ChatId = uint(chatID)
		return "GetChat", getChatReq, nil

	default:
		return "", nil, fmt.Errorf("unknown method: %s", dt.Method)
	}
}
