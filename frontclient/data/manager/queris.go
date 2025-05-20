package manager

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"orion/frontclient/utils/env"
	"orion/server/data/models"
)

var DB *gorm.DB

// init инициализирует подключение к базе данных.
// Функция загружает переменные окружения из файла ".env" и устанавливает соединение с базой данных PostgreSQL.
// При возникновении ошибок выполнение завершается с логированием ошибки.
func init() {
	var err error
	//err = godotenv.Load(".env")
	//

	//if err != nil {
	//	log.Fatalf("Some error occured. Err: %s", err)
	//}
	//

	DB, err = gorm.Open(postgres.Open(env.DatabaseUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
}

// GetUserByID возвращает пользователя по его ID.
// Выполняется поиск пользователя в базе данных по условию "id = ?". Результат упорядочивается по дате создания в порядке убывания.
//
// Параметры:
//   - userid: уникальный идентификатор пользователя.
//
// Возвращаемое значение:
//   - User: найденный пользователь (если не найден, вернется пустая структура User).
func GetUserByID(userid uint) models.User {
	var user models.User
	DB.Where("id = ?", userid).Find(&user).Order("created_at desc")
	return user
}
