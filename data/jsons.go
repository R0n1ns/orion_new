package data

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// общий тип запроса
type Querys struct {
	Method string      `json:"method"`
	Query  interface{} `json:"query"`
}

/*
{method:"",query:{}}
*/

// ----- от клиента -----

// RcvdMessage отправка сообщения
type RcvdMessage struct {
	ChatId  uint   `json:"chatId"`
	Message string `json:"message"`
}

/*
{method:"RcvdMessage",query:{chatId:1,message:"dsfds"}}
*/

// GetChats получить чаты пользователя
/*
{method:"GetChats",query:{}}
*/

// GetChat получить информацию о конкретном чате
/*
{method:"GetChat",query:{chatId:}}
*/
type GetChat struct {
	ChatId uint `json:"chatId"`
}

// ReadMessages прочитать сообщения от пользователя
/*
{method:"ReadMessages",query:{chatId:,userId:}}
*/

// GetUsers прочитать сообщения от пользователя
/*
{method:"GetUsers",query:{username:}}
*/

// CreateChat создать чат
/*
{method:"CreateChat",query:{user1:,user2:}}
*/

// ----- от сервера -----

// Отправление сообщения
/*
{method:"RetMessage",data:{fromChatID:,message:"",timestamp:}}
*/

//прочитан чат
/*
{method:"ReadedMessages",data:{chatId:}}
*/

//отправка чатов
/*
	{method:"GetChats",data:{[{"id":,"name":},]}}}
*/

//отправка юзеров
/*
	{method:"UsersAnsw",data:{[{"id":,"username":,"chat_id":},]}}}
*/

//отправка конкретного чата
/*
	{method:"GetChat",data:{[{"ID":,"ChannelID":,"UserFromID":,"Message":,"Edited":,"Readed":,"Date":},]}}
*/

type Default struct {
	Nothing string `json:"nothing"`
}

func HandleRequest(data []byte, userID uint) (string, interface{}, error) {
	var dt Querys
	json.Unmarshal(data, &dt)
	fmt.Println(dt)
	switch dt.Method {
	case "RcvdMessage":
		msg := RcvdMessage{}
		dat := dt.Query.(map[string]interface{})
		msg.Message = dat["message"].(string)
		msg.ChatId = uint(dat["chatId"].(float64))
		return "RcvdMessage", msg, nil
	case "ReadMessages":
		dat := dt.Query.(map[string]interface{})
		ReadMessages(dat["chatId"].(float64), dat["userId"].(string))
		return "ReadedMessages", dat["chatId"].(float64), nil
	case "GetChats":
		chats, err := GetChannels(userID)
		if err != nil {
			return "", nil, err
		}
		return "GetChats", chats, nil
	case "UpdateProfilePicture":
		data := dt.Query.(map[string]interface{})
		parts := strings.SplitN(data["imageData"].(string), ",", 2)
		if len(parts) != 2 {
			fmt.Errorf("некорректные данные изображения")
		}
		// Декодируем base64-часть
		imageBytes, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			fmt.Errorf("ошибка декодирования base64: %v", err)
		}
		// imageBytes содержит бинарные данные изображения
		hash := md5.Sum((imageBytes))
		imageHex := hex.EncodeToString(hash[:])

		AddHexPhoto(userID, imageHex)
		if _, err := os.Stat("images/" + imageHex + ".jpg"); errors.Is(err, os.ErrNotExist) {
			if err := ioutil.WriteFile(fmt.Sprint("images/"+imageHex+".jpg"), imageBytes, 0644); err != nil {
				fmt.Println("Ошибка при сохранении файла:", err)
			}
		}
		// Если требуется, "разжать" (декодировать) изображение можно просто повторно закодировав его в Base64.
		encodedImage := base64.StdEncoding.EncodeToString(imageBytes)
		// Формируем data URL для изображения
		profilePictureURL := "data:image/jpeg;base64," + encodedImage
		return "UpdateProfilePicture", profilePictureURL, nil
	case "GetUsers":
		dat := dt.Query.(map[string]interface{})
		users := SearchByUsername(dat["username"].(string))
		return "GetUsers", users, nil
	case "CreateChat":
		dat := dt.Query.(map[string]interface{})
		//fmt.Println(dat)
		chatname := dat["user1"].(string) + "--" + fmt.Sprint(dat["user2"].(float64))
		us1, _ := strconv.ParseUint(dat["user1"].(string), 10, 32)
		us2, _ := dat["user2"].(float64)
		chat, _ := CreateChat(uint(us1), uint(us2), chatname)

		//fmt.Println(chat)
		return "ChatCreated", *chat, nil
	case "GetChat":
		msg := GetChat{}
		dat := dt.Query.(map[string]interface{})
		msg.ChatId = uint(dat["chatId"].(float64))
		return "GetChat", msg, nil
	default:
		return "", nil, nil
	}
}

/* {"data":{"fromUserID":1,"message":"test","timestamp":"2025-01-08T15:30:47.8491888+03:00"},"method":"RcvdMessage"}
 */
