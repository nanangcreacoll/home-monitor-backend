package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UUID          uuid.UUID `gorm:"unique" json:"uuid" validate:"required,uuid"`
	Name          string    `gorm:"unique" json:"name"`
	MacAddress    string    `gorm:"unique" json:"mac_address"`
	UserCreatedID uint      `json:"user_created_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DeviceCreateRequest struct {
	Name       string `json:"name" binding:"required,lte=255"`
	MacAddress string `json:"mac_address" binding:"required,mac"`
}

type DeviceUpdateRequest struct {
	Name       string `json:"name" binding:"omitempty,lte=255"`
	MacAddress string `json:"mac_address" binding:"omitempty,mac"`
}

type DeviceDeleteRequest struct {
	UUID uuid.UUID `json:"uuid" binding:"required,uuid"`
}

type DeviceResponse struct {
	UUID          uuid.UUID `json:"uuid"`
	Name          string    `json:"name"`
	MacAddress    string    `json:"mac_address"`
	UserCreatedID uint      `json:"user_created_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DeviceListResponse struct {
	Devices []DeviceResponse `json:"devices"`
}

type DeviceListRequestParams struct {
	Length int  `form:"length" binding:"omitempty"`
	Latest bool `form:"latest" binding:"omitempty"`
}

type DeviceMeasurement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DeviceID    uint      `json:"device_id"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	CreatedAt   time.Time `json:"created_at"`
}

type DeviceMeasurementDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,dive,gt=0"`
}

type DeviceMeasurementPayload struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

type DeviceMeasurementListResponse struct {
	Measurements []DeviceMeasurement `json:"measurements"`
}

type DeviceMeasurementRequestParams struct {
	Length     int    `form:"length" binding:"omitempty"`
	DeviceUUID string `form:"device_uuid" binding:"omitempty,uuid"`
	Latest     bool   `form:"latest" binding:"omitempty"`
}

func (d *Device) BeforeCreate(tx *gorm.DB) (err error) {
	if d.UUID == uuid.Nil {
		d.UUID = uuid.New()
	}

	if d.Name == "" {
		return errors.New("device name is required")
	}

	if d.MacAddress == "" {
		return errors.New("mac address is required")
	}

	if d.UserCreatedID == 0 {
		return errors.New("user_created_id is required")
	}

	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()

	return nil
}

func (d *DeviceMeasurement) BeforeCreate(tx *gorm.DB) (err error) {
	if d.DeviceID == 0 {
		return errors.New("device_id is required")
	}

	d.CreatedAt = time.Now()

	return nil
}
