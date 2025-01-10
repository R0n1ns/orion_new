package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"orion/data"
	"strconv"
)

func GetChatsHandler(w http.ResponseWriter, r *http.Request) {
	userid, err := extractJWT(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	chats, err := data.GetChannels(userid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chatsJSON := make([]map[string]interface{}, len(chats))
	for i, chat := range chats {
		chatsJSON[i] = map[string]interface{}{
			"id":   chat.ID,
			"name": chat.Name,
		}
	}

	ret := map[string]interface{}{"chats": chatsJSON}

	//js, err := json.MarshalIndent(ret, "", "  ")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(string(js)) // выводим JSON в консоль

	//fmt.Println("ret", ret)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ret)
	return

}
func GetChatHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatid := vars["id"]
	id, _ := strconv.ParseUint(chatid, 10, 64)
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
		})
	}
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(masseges)
	return
}
