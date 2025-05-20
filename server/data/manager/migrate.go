package manager

import (
	"orion/server/data/models"
)

func Migrate() {

	DB.AutoMigrate(&models.User{}, &models.Message{}, &models.Channel{}, &models.Status{})
}
