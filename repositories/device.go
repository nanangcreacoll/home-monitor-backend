package repositories

import (
	"context"
	"home-monitor-backend/database"
	"home-monitor-backend/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceRepository interface {
	DeviceCreate(ctx context.Context, device *models.Device) error
	DeviceDelete(ctx context.Context, device *models.Device) error
	DeviceList(ctx context.Context, length int, latest bool) ([]models.Device, error)
	DeviceUpdate(ctx context.Context, device *models.Device) (models.Device, error)
	DeviceUpdateStatus(ctx context.Context, deviceID uint, status bool) error
	DeviceFindByName(ctx context.Context, name string) (*models.Device, error)
	DeviceFindByUUID(ctx context.Context, deviceUUID uuid.UUID) (*models.Device, error)
	DeviceFindByUserID(ctx context.Context, userID uint) ([]models.Device, error)
	DeviceFindByMacAddress(ctx context.Context, macAddress string) (*models.Device, error)
	DeviceMeasurementCreate(ctx context.Context, measurement *models.DeviceMeasurement) error
	DeviceMeasurementDelete(ctx context.Context, measurementID []uint) error
	DeviceMeasurements(ctx context.Context, length int, latest bool, startTime *string, endTime *string) ([]models.DeviceMeasurement, error)
	DeviceMeasurementsByDeviceID(ctx context.Context, length int, deviceID uint, latest bool, startTime *string, endTime *string) ([]models.DeviceMeasurement, error)
}

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository() DeviceRepository {
	return &deviceRepository{db: database.DB}
}

const defaultListLimit = 50

func (r *deviceRepository) DeviceCreate(ctx context.Context, device *models.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *deviceRepository) DeviceDelete(ctx context.Context, device *models.Device) error {
	return r.db.WithContext(ctx).Delete(device).Error
}

func (r *deviceRepository) DeviceFindByName(ctx context.Context, name string) (*models.Device, error) {
	var device models.Device
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) DeviceList(ctx context.Context, length int, latest bool) ([]models.Device, error) {
	var devices []models.Device
	query := r.db.WithContext(ctx)
	if latest {
		query = query.Order("created_at desc")
	}
	if length > 0 {
		query = query.Limit(length)
	} else {
		query = query.Limit(defaultListLimit)
	}

	if err := query.Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *deviceRepository) DeviceUpdate(ctx context.Context, device *models.Device) (models.Device, error) {
	var existingDevice models.Device
	if err := r.db.WithContext(ctx).First(&existingDevice, device.ID).Error; err != nil {
		return existingDevice, err
	}

	existingDevice = *device
	result := r.db.WithContext(ctx).Model(&existingDevice).Updates(existingDevice)
	return existingDevice, result.Error
}

func (r *deviceRepository) DeviceUpdateStatus(ctx context.Context, deviceID uint, status bool) error {
	return r.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", deviceID).Update("status", status).Error
}

func (r *deviceRepository) DeviceFindByUUID(ctx context.Context, deviceUUID uuid.UUID) (*models.Device, error) {
	var device models.Device
	if err := r.db.WithContext(ctx).Where("uuid = ?", deviceUUID).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) DeviceFindByUserID(ctx context.Context, userID uint) ([]models.Device, error) {
	var devices []models.Device
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *deviceRepository) DeviceFindByMacAddress(ctx context.Context, macAddress string) (*models.Device, error) {
	var device models.Device
	if err := r.db.WithContext(ctx).Where("mac_address = ?", macAddress).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) DeviceMeasurementCreate(ctx context.Context, measurement *models.DeviceMeasurement) error {
	return r.db.WithContext(ctx).Create(measurement).Error
}

func (r *deviceRepository) DeviceMeasurementDelete(ctx context.Context, measurementID []uint) error {
	return r.db.WithContext(ctx).Delete(&models.DeviceMeasurement{}, measurementID).Error
}

func (r *deviceRepository) deviceMeasurements(ctx context.Context, length int, latest bool, startTime *string, endTime *string, deviceID *uint) ([]models.DeviceMeasurement, error) {
	var measurements []models.DeviceMeasurement

	query := r.db.WithContext(ctx)

	if length <= 0 {
		length = defaultListLimit
	}
	query = query.Limit(length)

	if startTime != nil {
		start, err := time.Parse(time.RFC3339, *startTime)
		if err == nil {
			*startTime = start.Format(time.RFC3339)
		}
		query = query.Where("created_at >= ?", start)
	}

	if endTime != nil {
		end, err := time.Parse(time.RFC3339, *endTime)
		if err == nil {
			*endTime = end.Format(time.RFC3339)
		}
		query = query.Where("created_at <= ?", end)
	}

	if latest {
		query = query.Order("created_at desc")
	}

	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}

	return measurements, query.Find(&measurements).Error
}

func (r *deviceRepository) DeviceMeasurements(ctx context.Context, length int, latest bool, startTime *string, endTime *string) ([]models.DeviceMeasurement, error) {
	return r.deviceMeasurements(ctx, length, latest, startTime, endTime, nil)
}

func (r *deviceRepository) DeviceMeasurementsByDeviceID(ctx context.Context, length int, deviceID uint, latest bool, startTime *string, endTime *string) ([]models.DeviceMeasurement, error) {
	return r.deviceMeasurements(ctx, length, latest, startTime, endTime, &deviceID)
}
