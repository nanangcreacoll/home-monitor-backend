package seeders

import (
	"errors"

	"home-monitor-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func DeviceRun(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	devices := []models.Device{
		{
			Name:       "monitor-1",
			MacAddress: "00:00:00:00:00:01",
		},
		{
			Name:       "monitor-2",
			MacAddress: "00:00:00:00:00:02",
		},
		{
			Name:       "monitor-3",
			MacAddress: "00:00:00:00:00:03",
		},
	}

	return db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&devices).Error
}

func DeviceMeasurementRun(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	measurements := []models.DeviceMeasurement{
		{
			DeviceID:    1,
			Temperature: 25.5,
			Humidity:    60.0,
		},
		{
			DeviceID:    1,
			Temperature: 26.0,
			Humidity:    58.0,
		},
		{
			DeviceID:    1,
			Temperature: 24.5,
			Humidity:    65.0,
		},
	}

	return db.Create(&measurements).Error
}
