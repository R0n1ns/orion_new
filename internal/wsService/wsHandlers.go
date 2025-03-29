package wsService

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"orion/internal/data/managers/minio"
	data2 "orion/internal/data/managers/postgres"
	"orion/pkg/metrics"
	"strconv"
	"strings"
	"time"
)

type RequestHandler struct {
	Validator RequestValidator
}

func (h *RequestHandler) HandleRequest(data []byte, userID uint) (string, interface{}, error) {
	var dt Querys
	if err := json.Unmarshal(data, &dt); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	fmt.Println("Received request:", dt)

	// отправка метрик
	startTime := time.Now()
	defer func(method string, start time.Time) {
		elapsed := time.Since(start)
		// Отправляем время обработки в секундах с меткой, указывающей тип сообщения.
		metrics.MessageProcessingTime.WithLabelValues(method).Observe(elapsed.Seconds())
	}(dt.Method, startTime)

	// Валидируем данные перед обработкой
	if err := h.Validator.Validate(dt.Query.(map[string]interface{})); err != nil {
		return "", nil, err
	}

	// Основной процесс обработки запроса
	switch dt.Method {
	case "RcvdMessage":
		return h.handleRcvdMessage(dt.Query.(map[string]interface{}))
	case "ReadMessages":
		return h.handleReadMessages(dt.Query.(map[string]interface{}), userID)
	case "GetChats":
		return h.handleGetChats(userID)
	case "UpdateProfilePicture":
		return h.handleUpdateProfilePicture(dt.Query.(map[string]interface{}), userID)
	case "GetUsers":
		return h.handleGetUsers(dt.Query.(map[string]interface{}))
	case "CreateChat":
		return h.handleCreateChat(dt.Query.(map[string]interface{}), userID)
	case "GetChat":
		return h.handleGetChat(dt.Query.(map[string]interface{}))
	default:
		return "", nil, fmt.Errorf("unknown method: %s", dt.Method)
	}
}

func (h *RequestHandler) handleRcvdMessage(data map[string]interface{}) (string, interface{}, error) {
	msg := RcvdMessage{
		ChatId:  uint(data["chatId"].(float64)),
		Message: data["message"].(string),
	}
	// Сохранение сообщения в БД (если нужно)
	return "RcvdMessage", msg, nil
}

func (h *RequestHandler) handleReadMessages(data map[string]interface{}, userID uint) (string, interface{}, error) {
	// Обработка запроса на пометку сообщений как прочитанных.
	chatID := uint(data["chatId"].(float64))
	userIDStr := data["userId"].(string)
	data2.ReadMessages(float64(chatID), userIDStr)
	return "ReadedMessages", chatID, nil
}

func (h *RequestHandler) handleGetChats(userID uint) (string, interface{}, error) {
	chats, err := data2.GetChannels(userID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get channels: %w", err)
	}
	return "GetChats", chats, nil
}

func (h *RequestHandler) handleUpdateProfilePicture(data map[string]interface{}, userID uint) (string, interface{}, error) {
	imageData := data["imageData"].(string)
	parts := strings.SplitN(imageData, ",", 2)
	if len(parts) != 2 {
		return "", nil, errors.New("invalid image data format")
	}

	// Декодирование base64-части изображения.
	imageBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("base64 decoding error: %w", err)
	}

	// Обновление данных в БД и загрузка в MinIO
	hash := md5.Sum(imageBytes)
	imageHex := hex.EncodeToString(hash[:])
	data2.AddHexPhoto(userID, imageHex)

	objectName := imageHex + ".jpg"
	err = minio.MinioMgr.UploadImage(context.Background(), objectName, imageBytes, "image/jpeg")
	if err != nil {
		return "", nil, fmt.Errorf("failed to upload image to minio: %w", err)
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageBytes)
	profilePictureURL := "data:image/jpeg;base64," + encodedImage
	return "UpdateProfilePicture", profilePictureURL, nil
}

func (h *RequestHandler) handleGetUsers(data map[string]interface{}) (string, interface{}, error) {
	username := data["username"].(string)
	users := data2.SearchByUsername(username)
	return "GetUsers", users, nil
}

func (h *RequestHandler) handleCreateChat(data map[string]interface{}, userID uint) (string, interface{}, error) {
	user1Str := data["user1"].(string)
	user2Float := data["user2"].(float64)
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
}

func (h *RequestHandler) handleGetChat(data map[string]interface{}) (string, interface{}, error) {
	chatID := uint(data["chatId"].(float64))
	getChatReq := GetChat{ChatId: chatID}
	return "GetChat", getChatReq, nil
}
