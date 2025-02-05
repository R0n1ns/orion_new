package data

import (
	"encoding/base64"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
)

var DB *gorm.DB

// init инициализирует подключение к базе данных.
// Функция загружает переменные окружения из файла ".env" и устанавливает соединение с базой данных PostgreSQL.
// При возникновении ошибок выполнение завершается с логированием ошибки.
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
	// Примеры создания пользователей и чатов (закомментировано):
	//DB.Create(&User{Mail: "test2@dsfds.txt", Password: "test2", UserName: "test2", LastOnline: time.Now()})
	//DB.Create(&User{Mail: "test3@dsfds.txt", Password: "test3", UserName: "test3", LastOnline: time.Now()})
	//CreateChat(1, 2, "chat2")
	//CreateChat(1, 3, "chat3")
	//CreateChat(1, 1)
}

// GetUserByID возвращает пользователя по его ID.
// Выполняется поиск пользователя в базе данных по условию "id = ?". Результат упорядочивается по дате создания в порядке убывания.
//
// Параметры:
//   - userid: уникальный идентификатор пользователя.
//
// Возвращаемое значение:
//   - User: найденный пользователь (если не найден, вернется пустая структура User).
func GetUserByID(userid uint) User {
	var user User
	DB.Where("id = ?", userid).Find(&user).Order("created_at desc")
	return user
}

// GetChatIDForUsers ищет личный чат между двумя пользователями по их ID.
//
// Для случая, когда оба ID совпадают (чат с самим собой), функция ищет чат, в котором присутствует только один пользователь с данным ID.
// Для разных ID функция ищет чат, в котором присутствуют оба пользователя, и общее количество участников равно двум.
//
// Параметры:
//   - user1ID: ID первого пользователя.
//   - user2ID: ID второго пользователя.
//
// Возвращаемое значение:
//   - int: ID найденного чата. Если чат не найден или произошла ошибка, возвращается -1.
func GetChatIDForUsers(user1ID, user2ID uint) int {
	// Если запрашивается чат самого с собой:
	if user1ID == user2ID {
		var channelIDs []int
		err := DB.
			Table("channels").
			Joins("JOIN user_channels ON user_channels.channel_id = channels.id").
			Group("channels.id").
			Having("COUNT(user_channels.user_id) = ? AND SUM(CASE WHEN user_channels.user_id = ? THEN 1 ELSE 0 END) = ?", 1, user1ID, 1).
			Pluck("channels.id", &channelIDs).Error

		if err != nil {
			fmt.Println(err)
			return -1
		}

		if len(channelIDs) == 0 {
			// Чат самого с собой не найден
			return -1
		}
		// Если найдено несколько, возвращаем первый найденный
		return channelIDs[0]
	}

	// Обработка обычного личного чата между двумя разными пользователями:
	var channelIDs []int
	err := DB.
		Table("channels").
		Joins("JOIN user_channels ON user_channels.channel_id = channels.id").
		Group("channels.id").
		Having("COUNT(user_channels.user_id) = ? AND SUM(CASE WHEN user_channels.user_id IN (?, ?) THEN 1 ELSE 0 END) = ?", 2, user1ID, user2ID, 2).
		Pluck("channels.id", &channelIDs).Error

	if err != nil {
		fmt.Println(err)
		return -1
	}

	if len(channelIDs) == 0 {
		// Чат не найден
		return -1
	}
	// Если найдено несколько совпадений, возвращаем первый найденный
	return channelIDs[0]
}

// ReadMessages помечает сообщения в указанном чате как прочитанные для заданного пользователя.
//
// Функция обновляет все сообщения в канале с идентификатором chatid,
// где пользователь не является отправителем и флаг readed равен false.
//
// Параметры:
//   - chatid: идентификатор чата (канала).
//   - userid: идентификатор пользователя, для которого устанавливается флаг прочтения.
func ReadMessages(chatid float64, userid string) {
	DB.Model(&Message{}).Where("user_id != ? and channel_id = ? and readed = false", userid, chatid).Update("readed", true)
}

// GetChatByID возвращает информацию о чате (канале) по его ID.
//
// Параметры:
//   - chatid: уникальный идентификатор чата.
//
// Возвращаемое значение:
//   - Channel: найденный чат. Если чат не найден, возвращается пустая структура Channel.
func GetChatByID(chatid uint) Channel {
	var chat Channel
	DB.Where("id = ?", chatid).Find(&chat).Order("created_at desc")
	return chat
}

// GetChannels возвращает список всех каналов, к которым принадлежит пользователь.
//
// Параметры:
//   - userid: уникальный идентификатор пользователя.
//
// Возвращаемые значения:
//   - []Channel: слайс с каналами, связанными с пользователем.
//   - error: ошибка, если произошла неудача при получении данных.
func GetChannels(userid uint) ([]Channel, error) {
	var channels []Channel
	err := DB.Model(&User{ID: userid}).Association("Channels").Find(&channels)
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// GetChanMassages возвращает список сообщений, принадлежащих каналу, отсортированных по времени отправки (возрастание).
//
// Параметры:
//   - chanid: уникальный идентификатор канала.
//
// Возвращаемые значения:
//   - []Message: слайс сообщений канала.
//   - error: ошибка, если произошла неудача при получении данных.
func GetChanMassages(chanid uint) ([]Message, error) {
	var message []Message
	err := DB.Model(&Channel{ID: chanid}).Association("Messages").Find(&message)
	if err != nil {
		return nil, err
	}
	sort.Slice(message, func(i, j int) bool {
		return message[i].Timestamp.Before(message[j].Timestamp)
	})
	return message, nil
}

// AddMessage добавляет новое сообщение в указанный чат.
//
// Параметры:
//   - froid: идентификатор пользователя (отправителя сообщения).
//   - chaid: идентификатор чата (канала), куда отправляется сообщение.
//   - message: текст сообщения.
//
// В случае ошибки при сохранении сообщения в базу данных, ошибка логируется.
func AddMessage(froid uint, chaid uint, message string) {
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

// AddHexPhoto обновляет фотографию профиля пользователя.
// Фотография представлена в виде строки с шестнадцатеричным кодом (hex).
//
// Параметры:
//   - userid: уникальный идентификатор пользователя.
//   - hex: строка с шестнадцатеричным представлением изображения.
//
// При возникновении ошибки обновления записи, ошибка логируется.
func AddHexPhoto(userid uint, hex string) {
	err := DB.Model(&User{}).Where("id = ?", userid).Update("profile_picture", hex)
	if err != nil {
		log.Printf("Some error occured. Err: %s", err)
	}
}

// CreateChat создаёт новый приватный чат между двумя пользователями, если такой чат ещё не существует.
// Если чат с указанным именем уже существует, функция возвращает его.
//
// Параметры:
//   - userID1: идентификатор первого пользователя (инициатора чата).
//   - userID2: идентификатор второго пользователя.
//   - channelName: уникальное имя канала (чата).
//
// Возвращаемые значения:
//   - *Channel: указатель на созданный или существующий чат.
//   - error: ошибка, если один из пользователей не найден или не удалось создать чат.
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

// SearchByUsername выполняет поиск пользователей по началу их имени пользователя.
// Используется оператор LIKE с подстановкой, чтобы найти всех пользователей, имя которых начинается с заданной строки.
//
// Параметры:
//   - username: начальная часть имени пользователя для поиска.
//
// Возвращаемое значение:
//   - []User: слайс найденных пользователей.
func SearchByUsername(username string) []User {
	var user []User
	query := username + "%"
	DB.Where("user_name LIKE ?", query).Find(&user)
	return user
}

// GetUsersInChat возвращает список пользователей, участвующих в указанном чате.
//
// Параметры:
//   - chatid: уникальный идентификатор чата (канала).
//
// Возвращаемые значения:
//   - []User: слайс пользователей, находящихся в чате.
//   - error: ошибка, если не удалось получить список пользователей.
func GetUsersInChat(chatid uint) ([]User, error) {
	var users []User
	err := DB.Model(&Channel{ID: chatid}).Association("Users").Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// IfReadedChat проверяет, есть ли в указанном чате непрочитанные сообщения для пользователя.
//
// Функция подсчитывает количество сообщений в канале с chatid,
// которые были отправлены другими пользователями, не прочитаны (readed = false)
// и не удалены (deleted_at IS NULL).
//
// Параметры:
//   - chatid: уникальный идентификатор чата.
//   - userID: идентификатор пользователя, для которого производится проверка.
//
// Возвращаемое значение:
//   - bool: возвращает true, если все сообщения прочитаны (или их нет), иначе false.
func IfReadedChat(chatid uint, userID uint) bool {
	var count int64
	err := DB.Model(&Message{}).
		Where("\"messages\".\"channel_id\" = ?", chatid).
		Where("\"messages\".\"user_id\" != ?", userID).
		Where("\"messages\".\"readed\" = false").
		Where("\"messages\".\"deleted_at\" IS NULL").
		Count(&count).Error

	if err != nil || count > 0 {
		return false // есть непрочитанные сообщения
	}
	return true // нет непрочитанных сообщений
}

// GetPhoto возвращает фотографию в виде data URL для изображения, если файл существует.
// Файл ищется по пути "images/{hash}.jpg". Если файл не найден, возвращается строка "none".
//
// Параметры:
//   - hash: строка-хэш, которая используется для формирования имени файла.
//
// Возвращаемое значение:
//   - string: data URL изображения или "none", если изображение не найдено.
func GetPhoto(hash string) string {
	file, err := ioutil.ReadFile(fmt.Sprint("images/" + hash + ".jpg"))
	if err != nil {
		//fmt.Println("Ошибка при открытии файла:", err)
		return "none"
	}
	encodedImage := base64.StdEncoding.EncodeToString(file)
	// Формируем data URL для изображения
	profilePictureURL := "data:image/jpeg;base64," + encodedImage
	return profilePictureURL
}
