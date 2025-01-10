package data

import (
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

var DB *gorm.DB

func init() {
	var err error
	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	url := os.Getenv("DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	//DB.Create(&User{Mail: "test2@dsfds.txt", Password: "test2", UserName: "test2", LastOnline: time.Now()})
	//DB.Create(&User{Mail: "test3@dsfds.txt", Password: "test3", UserName: "test3", LastOnline: time.Now()})
	//CreateChat(1, 2, "chat2")
	//CreateChat(1, 3, "chat3")

	//CreateChat(1, 1)
}
func GetUserByID(userid uint) User {
	var user User
	DB.Where("id = ?", userid).Find(&user).Order("created_at desc")
	return user
}

func GetChannels(userid uint) ([]Channel, error) {
	var channels []Channel
	err := DB.Model(&User{ID: userid}).Association("Channels").Find(&channels)
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func GetChanMassages(chanid uint) ([]Message, error) {
	var message []Message
	err := DB.Model(&Channel{ID: chanid}).Association("Messages").Find(&message)
	if err != nil {
		return nil, err
	}
	return message, nil
}
func AddMessage(froid uint, chaid uint, message string) {
	//var user User
	//if err := DB.First(&user, froid).Error; err != nil {
	//	fmt.Printf("user not found: %w", err)
	//}
	//var chanel Channel
	//if err := DB.First(&chanel, chaid).Error; err != nil {
	//	fmt.Printf("channen not found: %w", err)
	//}
	mess := Message{
		ChannelID: chaid,
		UserID:    froid,
		Content:   message,
		Timestamp: time.Now(),
	}
	err := DB.Create(&mess).Error
	if err != nil {
		log.Printf("Some error occured. Err: %s", err)
	}
}
func CreateChat(userID1, userID2 uint, channelName string) (*Channel, error) {
	// Проверяем, существуют ли оба пользователя
	var user1, user2 User
	if err := DB.First(&user1, userID1).Error; err != nil {
		return nil, fmt.Errorf("user 1 not found: %w", err)
	}
	if err := DB.First(&user2, userID2).Error; err != nil {
		return nil, fmt.Errorf("user 2 not found: %w", err)
	}

	// Создаём уникальное имя для канала
	//channelName := fmt.Sprintf("chat_%d_%d", userID1, userID2)

	// Проверяем, существует ли уже такой чат
	var existingChannel Channel
	if err := DB.Where("name = ?", channelName).First(&existingChannel).Error; err == nil {
		return &existingChannel, nil // Чат уже существует
	}

	// Создаём новый канал
	chat := Channel{
		Name:        channelName,
		Description: fmt.Sprintf("Private chat between user %d and user %d", userID1, userID2),
		IsPrivate:   true,
		CreatorID:   userID1,
		Users:       []User{user1, user2},
	}
	if err := DB.Create(&chat).Error; err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	return &chat, nil
}

func GetUsersInChat(chatid uint) ([]User, error) {
	var users []User
	err := DB.Model(&Channel{ID: chatid}).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil

}

//func main() {
//	Migrate()
//}
/*

 */
