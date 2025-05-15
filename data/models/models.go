package models

import (
	"gorm.io/gorm"
	"time"
)

// User представляет модель пользователя в системе.
//
// Поля структуры:
//   - ID: Уникальный идентификатор пользователя (primary key, автоинкремент).
//   - Mail: Email пользователя, используется в качестве логина (уникальное, не может быть пустым).
//   - UserName: Уникальное имя пользователя (обязательное поле).
//   - Password: Хэш пароля пользователя (обязательное поле).
//   - IsBlocked: Флаг, указывающий, заблокирован ли пользователь (по умолчанию false).
//   - LastOnline: Время последней активности пользователя (обязательное поле).
//   - ProfilePicture: Ссылка или код изображения профиля (текстовое поле, по умолчанию пустое).
//   - Bio: Биография пользователя, ограниченная 255 символами (по умолчанию пустая).
//
// Связи:
//   - Channels: Множество каналов, в которых состоит пользователь (многие ко многим через таблицу user_channels).
//   - Statuses: Множество статусов пользователя в каналах (многие ко многим через таблицу user_statuses).
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

	Channels     []Channel `gorm:"many2many:user_channels;"` // Множество каналов, в которых состоит пользователь
	Statuses     []Status  `gorm:"many2many:user_statuses;"` // Множество статусов пользователя в каналах
	BlockedUsers []User    `gorm:"many2many:user_blocks;joinForeignKey:BlockedID;joinReferences:BlockerID"`
}

// Channel представляет модель канала (чата) в системе.
//
// Поля структуры:
//   - ID: Уникальный идентификатор канала (primary key, автоинкремент).
//   - Name: Уникальное имя канала (обязательное поле).
//   - Description: Описание канала (текстовое поле, по умолчанию пустое).
//   - IsPrivate: Флаг приватности канала (по умолчанию false).
//   - CreatorID: Идентификатор пользователя, создавшего канал (обязательное поле).
//
// Связи:
//   - Creator: Пользователь, создавший канал (отношение «один к одному», внешний ключ – CreatorID).
//   - Users: Пользователи, участвующие в канале (отношение многие ко многим через таблицу user_channels).
//   - Messages: Сообщения, принадлежащие каналу (при удалении канала связанные сообщения также удаляются – CASCADE).
//   - Statuses: Статусы, связанные с каналом (один ко многим, внешний ключ ChannelID).
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

// Status представляет статус (роль) пользователя в канале.
//
// Поля структуры:
//   - ID: Уникальный идентификатор статуса (primary key, автоинкремент).
//   - Name: Название статуса (обязательное поле).
//   - ChannelID: Идентификатор канала, к которому принадлежит статус (обязательное поле).
//
// Привилегии (флаги), определяющие права:
//   - EditingPriv: Право редактировать сообщения (по умолчанию false).
//   - DeletionPriv: Право удалять сообщения (по умолчанию false).
//   - WritingPriv: Право отправлять сообщения (по умолчанию true).
//   - AdminPriv: Административные права (по умолчанию false).
//   - ManagingUsersPriv: Право управления пользователями (по умолчанию false).
//   - PinningMessagesPriv: Право закреплять сообщения (по умолчанию false).
//   - ReadingPriv: Право читать сообщения (по умолчанию true).
//   - InvitingPriv: Право приглашать пользователей (по умолчанию false).
//   - ViewingAnalyticsPriv: Право просматривать аналитику (по умолчанию false).
//   - ChannelEditPriv: Право редактировать канал (по умолчанию false).
//
// Связь:
//   - Channel: Канал, к которому принадлежит статус.
type Status struct {
	gorm.Model
	ID        uint    `gorm:"primaryKey;autoIncrement"` // Уникальный ID статуса
	Name      string  `gorm:"not null"`                 // Название статуса
	Channel   Channel // Канал, к которому принадлежит статус
	ChannelID uint    `gorm:"not null"` // ID канала, которому принадлежит статус

	// Привилегии пользователя в канале
	EditingPriv          bool `gorm:"default:false"` // Право редактировать сообщения
	DeletionPriv         bool `gorm:"default:false"` // Право удалять сообщения
	WritingPriv          bool `gorm:"default:true"`  // Право отправлять сообщения
	AdminPriv            bool `gorm:"default:false"` // Админские права
	ManagingUsersPriv    bool `gorm:"default:false"` // Право управления пользователями
	PinningMessagesPriv  bool `gorm:"default:false"` // Право закрепления сообщений
	ReadingPriv          bool `gorm:"default:true"`  // Право чтения сообщений
	InvitingPriv         bool `gorm:"default:false"` // Право приглашать пользователей
	ViewingAnalyticsPriv bool `gorm:"default:false"` // Право просмотра аналитики
	ChannelEditPriv      bool `gorm:"default:false"` // Право редактирования канала
}

// Message представляет сообщение, отправленное в канале.
//
// Поля структуры:
//   - ID: Уникальный идентификатор сообщения (primary key, автоинкремент).
//   - ChannelID: Идентификатор канала, к которому принадлежит сообщение (обязательное поле).
//   - UserID: Идентификатор пользователя, отправившего сообщение (обязательное поле).
//   - Content: Текстовое содержимое сообщения (обязательное поле).
//   - Timestamp: Время отправки сообщения (обязательное поле).
//   - Edited: Флаг, указывающий, было ли сообщение изменено (по умолчанию false).
//   - Readed: Флаг, указывающий, прочитано ли сообщение (по умолчанию false).
//
// Связи:
//   - Channel: Канал, к которому принадлежит сообщение (внешний ключ – ChannelID).
//   - User: Пользователь, отправивший сообщение (внешний ключ – UserID).
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
