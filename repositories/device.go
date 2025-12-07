package repositories

import (
	"home-monitor-backend/database"
	"home-monitor-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceRepository interface {
	DeviceCreate(device *models.Device) error
	DeviceDelete(device *models.Device) error
	DeviceList(length int, latest bool) ([]models.Device, error)
	DeviceUpdate(device *models.Device) (models.Device, error)
	DeviceFindByUUID(deviceUUID uuid.UUID) (*models.Device, error)
	DeviceFindByUserID(userID uint) ([]models.Device, error)
	DeviceFindByMacAddress(macAddress string) (*models.Device, error)
	DeviceMeasurementCreate(measurement *models.DeviceMeasurement) error
	DeviceMeasurementLatestByDeviceID(deviceID uint, length int) ([]models.DeviceMeasurement, error)
	DeviceMeasurementLatest(length int) ([]models.DeviceMeasurement, error)
	DeviceMeasurementByTimeRange(startTime, endTime string) ([]models.DeviceMeasurement, error)
	DeviceMeasurementByTimeRangeDeviceID(startTime, endTime string, deviceID uint) ([]models.DeviceMeasurement, error)
}

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository() DeviceRepository {
	return &deviceRepository{db: database.DB}
}

func (r *deviceRepository) DeviceCreate(device *models.Device) error {
	return r.db.Create(device).Error
}

func (r *deviceRepository) DeviceDelete(device *models.Device) error {
	return r.db.Delete(device).Error
}

func (r *deviceRepository) DeviceList(length int, latest bool) ([]models.Device, error) {
	var devices []models.Device
	query := r.db
	if latest {
		query = query.Order("created_at desc")
	}
	if length > 0 {
		query = query.Limit(length)
	}
	if err := query.Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *deviceRepository) DeviceUpdate(device *models.Device) (models.Device, error) {
	var existingDevice models.Device
	if err := r.db.First(&existingDevice, device.ID).Error; err != nil {
		return existingDevice, err
	}

	existingDevice = *device
	result := r.db.Model(&existingDevice).Updates(existingDevice)
	return existingDevice, result.Error
}

func (r *deviceRepository) DeviceFindByUUID(deviceUUID uuid.UUID) (*models.Device, error) {
	var device models.Device
	if err := r.db.Where("uuid = ?", deviceUUID).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) DeviceFindByUserID(userID uint) ([]models.Device, error) {
	var devices []models.Device
	if err := r.db.Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *deviceRepository) DeviceFindByMacAddress(macAddress string) (*models.Device, error) {
	var device models.Device
	if err := r.db.Where("mac_address = ?", macAddress).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) DeviceMeasurementCreate(measurement *models.DeviceMeasurement) error {
	return r.db.Create(measurement).Error
}

func (r *deviceRepository) DeviceMeasurementLatestByDeviceID(deviceID uint, length int) ([]models.DeviceMeasurement, error) {
	var measurements []models.DeviceMeasurement
	if err := r.db.Where("device_id = ?", deviceID).Order("created_at desc").Limit(length).Find(&measurements).Error; err != nil {
		return nil, err
	}
	return measurements, nil
}

func (r *deviceRepository) DeviceMeasurementLatest(length int) ([]models.DeviceMeasurement, error) {
	var measurements []models.DeviceMeasurement
	if err := r.db.Order("created_at desc").Limit(length).Find(&measurements).Error; err != nil {
		return nil, err
	}
	return measurements, nil
}

func (r *deviceRepository) DeviceMeasurementByTimeRange(startTime, endTime string) ([]models.DeviceMeasurement, error) {
	var measurements []models.DeviceMeasurement
	if err := r.db.Where("created_at BETWEEN ? AND ?", startTime, endTime).Order("created_at desc").Find(&measurements).Error; err != nil {
		return nil, err
	}
	return measurements, nil
}

func (r *deviceRepository) DeviceMeasurementByTimeRangeDeviceID(startTime, endTime string, deviceID uint) ([]models.DeviceMeasurement, error) {
	var measurements []models.DeviceMeasurement
	if err := r.db.Where("device_id = ? AND created_at BETWEEN ? AND ?", deviceID, startTime, endTime).Order("created_at desc").Find(&measurements).Error; err != nil {
		return nil, err
	}
	return measurements, nil
}
