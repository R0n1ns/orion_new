package migrate

import (
	"orion/internal/data/managers/postgres"
	"orion/internal/data/models"
)

func Migrate() {

	postgres.DB.AutoMigrate(&models.User{}, &models.Message{}, &models.Channel{}, &models.Status{})
}
