package data

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	ID             uint      `gorm:"primaryKey;autoIncrement"`     // Уникальный ID пользователя
	Mail           string    `gorm:"unique;not null"`              // Email пользователя, используется как логин (уникальный)
	UserName       string    `gorm:"unique;not null"`              // Уникальное имя пользователя
	Password       string    `gorm:"not null"`                     // Хэш пароля
	IsBlocked      bool      `gorm:"default:false"`                // Заблокирован ли пользователь
	LastOnline     time.Time `gorm:"not null"`                     // Последнее время активности пользователя
	ProfilePicture string    `gorm:"type:text;default:''"`         // Ссылка на картинку профиля
	Bio            string    `gorm:"type:varchar(255);default:''"` // Био пользователя

	Channels []Channel `gorm:"many2many:user_channels;"` // Множество каналов
	Statuses []Status  `gorm:"many2many:user_statuses;"` // Множество статусов в каналах
}

type Channel struct {
	gorm.Model
	ID          uint      `gorm:"primaryKey;autoIncrement"`    // Уникальный ID канала
	Name        string    `gorm:"unique;not null"`             // Уникальное имя канала
	Description string    `gorm:"type:text;default:''"`        // Описание канала
	IsPrivate   bool      `gorm:"default:false"`               // Приватность канала
	CreatorID   uint      `gorm:"not null"`                    // ID создателя канала
	Creator     User      `gorm:"foreignKey:CreatorID"`        // Связь с создателем канала
	Users       []User    `gorm:"many2many:user_channels;"`    // Пользователи, участвующие в канале
	Messages    []Message `gorm:"constraint:OnDelete:CASCADE"` // Сообщения канала
	Statuses    []Status  `gorm:"foreignKey:ChannelID"`        // Статусы, связанные с каналом
}

type Status struct {
	gorm.Model
	ID        uint    `gorm:"primaryKey;autoIncrement"` // Уникальный ID статуса
	Name      string  `gorm:"not null"`                 // Название статуса
	Channel   Channel // Канал, к которому принадлежит статус
	ChannelID uint    `gorm:"not null"` // ID канала, которому принадлежит статус

	// Привилегии
	EditingPriv          bool `gorm:"default:false"` // Право редактировать сообщения
	DeletionPriv         bool `gorm:"default:false"` // Право удалять сообщения
	WritingPriv          bool `gorm:"default:true"`  // Право отправлять сообщения
	AdminPriv            bool `gorm:"default:false"` // Админские права
	ManagingUsersPriv    bool `gorm:"default:false"` // Управление пользователями
	PinningMessagesPriv  bool `gorm:"default:false"` // Закрепление сообщений
	ReadingPriv          bool `gorm:"default:true"`  // Чтение сообщений
	InvitingPriv         bool `gorm:"default:false"` // Приглашение пользователей
	ViewingAnalyticsPriv bool `gorm:"default:false"` // Просмотр аналитики
	ChannelEditPriv      bool `gorm:"default:false"` // Редактирование канала
}

type Message struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey;autoIncrement"` // Уникальный ID сообщения
	ChannelID uint      `gorm:"not null"`                 // ID канала, которому принадлежит сообщение
	Channel   Channel   `gorm:"foreignKey:ChannelID"`     // Связь с каналом
	UserID    uint      `gorm:"not null"`                 // ID пользователя, отправившего сообщение
	User      User      `gorm:"foreignKey:UserID"`        // Связь с пользователем
	Content   string    `gorm:"type:text;not null"`       // Содержимое сообщения
	Timestamp time.Time `gorm:"not null"`                 // Время отправки сообщения
	Edited    bool      `gorm:"default:false"`            // Было ли сообщение изменено
	Readed    bool      `gorm:"default:false"`            // Было ли сообщение прочитано
}
