package data

import (
	"encoding/json"
	"fmt"
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
	case "GetUsers":
		dat := dt.Query.(map[string]interface{})
		users := SearchByUsername(dat["username"].(string))
		return "GetUsers", users, nil
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
