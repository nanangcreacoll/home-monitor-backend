package seeders

import (
	"errors"

	"home-monitor-backend/models"

	"gorm.io/gorm"
)

func UserRun(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	user := models.User{
		Username: "admin",
		Password: "password",
	}

	if err := db.Create(&user).Error; err != nil {
		return errors.New("failed to seed user: " + err.Error())
	}

	return nil
}
