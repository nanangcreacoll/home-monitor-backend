package seeders

import (
	"errors"

	"home-monitor-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UserRun(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	users := []models.User{
		{UUID: uuid.New(), Username: "admin", Password: "password", Role: models.UserRoleAdmin},
		{UUID: uuid.New(), Username: "admin2", Password: "password2", Role: models.UserRoleAdmin},
		{UUID: uuid.New(), Username: "user", Password: "password", Role: models.UserRoleUser},
		{UUID: uuid.New(), Username: "user2", Password: "password2", Role: models.UserRoleUser},
	}

	for i := range users {
		if users[i].UUID == uuid.Nil {
			users[i].UUID = uuid.New()
		}
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoNothing: true,
	}).Select("UUID", "Username", "Password", "Role").Create(&users).Error; err != nil {
		return errors.New("failed to seed users: " + err.Error())
	}

	return nil
}
