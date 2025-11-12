package services

import (
	"errors"
	"home-monitor-backend/models"
	"home-monitor-backend/repositories"
	"net/http"

	"github.com/google/uuid"
)

type DeviceService interface {
	DeviceRegister(device *models.Device, userUUID uuid.UUID) (*models.Device, int, error)
	DeviceDelete(deviceUUID uuid.UUID, userUUID uuid.UUID) (int, error)
	DeviceList(userUUID uuid.UUID, length int, latest bool) ([]models.Device, int, error)
}

type deviceService struct {
	deviceRepo repositories.DeviceRepository
	userRepo   repositories.UserRepository
}

func NewDeviceService(deviceRepo repositories.DeviceRepository, userRepo repositories.UserRepository) DeviceService {
	return &deviceService{deviceRepo: deviceRepo, userRepo: userRepo}
}

func (s *deviceService) DeviceRegister(device *models.Device, userUUID uuid.UUID) (*models.Device, int, error) {
	user, err := s.userRepo.UserFindByUUID(userUUID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return nil, http.StatusForbidden, errors.New("only admin can create new devices")
	}

	existingDevice, _ := s.deviceRepo.DeviceFindByMacAddress(device.MacAddress)
	if existingDevice != nil {
		return nil, http.StatusConflict, errors.New("device with this MAC address already exists")
	}

	device.UserCreatedID = user.ID
	if err := s.deviceRepo.DeviceCreate(device); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return device, http.StatusCreated, nil
}

func (s *deviceService) DeviceDelete(deviceUUID uuid.UUID, userUUID uuid.UUID) (int, error) {
	user, err := s.userRepo.UserFindByUUID(userUUID)
	if err != nil {
		return http.StatusNotFound, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return http.StatusForbidden, errors.New("only admin can delete devices")
	}

	device, err := s.deviceRepo.DeviceFindByUUID(deviceUUID)
	if err != nil {
		return http.StatusNotFound, errors.New("device not found")
	}

	measurements, err := s.deviceRepo.DeviceMeasurementLatestByDeviceID(device.ID, 1)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if len(measurements) > 0 {
		return http.StatusForbidden, errors.New("cannot delete device with existing measurements")
	}

	if err := s.deviceRepo.DeviceDelete(device); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (s *deviceService) DeviceList(userUUID uuid.UUID, length int, latest bool) ([]models.Device, int, error) {
	_, err := s.userRepo.UserFindByUUID(userUUID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("user not found")
	}

	devices, err := s.deviceRepo.DeviceList(length, latest)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return devices, http.StatusOK, nil
}
