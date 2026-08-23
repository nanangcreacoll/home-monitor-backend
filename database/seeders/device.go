package seeders

import (
	"errors"
	"time"

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

	var device models.Device
	if err := db.First(&device).Error; err != nil {
		return errors.New("no devices found for seeding measurements")
	}

	measurements := []models.DeviceMeasurement{
		{
			DeviceID:    device.ID,
			Temperature: 25.5,
			Humidity:    60.0,
			CreatedAt:   time.Now().Add(-10 * time.Minute),
		},
		{
			DeviceID:    device.ID,
			Temperature: 26.0,
			Humidity:    58.0,
			CreatedAt:   time.Now().Add(-5 * time.Minute),
		},
		{
			DeviceID:    device.ID,
			Temperature: 24.5,
			Humidity:    65.0,
			CreatedAt:   time.Now(),
		},
	}

	return db.Create(&measurements).Error
}
