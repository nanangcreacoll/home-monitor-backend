package database

import (
	"errors"

	"home-monitor-backend/database/seeders"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	if err := seeders.UserRun(db); err != nil {
		return errors.New("failed to run user seeder: " + err.Error())
	}

	return nil
}

func Seed() error {
	if DB == nil {
		return errors.New("database not initialized")
	}
	return Run(DB)
}
