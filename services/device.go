package services

import (
	"errors"
	"home-monitor-backend/models"
	"home-monitor-backend/repositories"
	"net/http"

	"github.com/google/uuid"
)

type DeviceService interface {
	DeviceRegister(userUUID uuid.UUID, device *models.Device) (*models.Device, int, error)
	DeviceDelete(userUUID uuid.UUID, deviceUUID uuid.UUID) (int, error)
	DeviceList(userUUID uuid.UUID, length int, latest bool) ([]models.Device, int, error)
	DeviceProfile(userUUID uuid.UUID, deviceUUID uuid.UUID) (*models.Device, int, error)
	DeviceUpdate(userUUID uuid.UUID, deviceUUID uuid.UUID, deviceUpdate *models.DeviceUpdateRequest) (*models.Device, int, error)
}

type deviceService struct {
	deviceRepo repositories.DeviceRepository
	userRepo   repositories.UserRepository
}

func NewDeviceService(deviceRepo repositories.DeviceRepository, userRepo repositories.UserRepository) DeviceService {
	return &deviceService{deviceRepo: deviceRepo, userRepo: userRepo}
}

func (s *deviceService) DeviceRegister(userUUID uuid.UUID, device *models.Device) (*models.Device, int, error) {
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

func (s *deviceService) DeviceDelete(userUUID uuid.UUID, deviceUUID uuid.UUID) (int, error) {
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

func (s *deviceService) DeviceProfile(userUUID uuid.UUID, deviceUUID uuid.UUID) (*models.Device, int, error) {
	_, err := s.userRepo.UserFindByUUID(userUUID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("user not found")
	}

	device, err := s.deviceRepo.DeviceFindByUUID(deviceUUID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("device not found")
	}
	return device, http.StatusOK, nil
}

func (s *deviceService) DeviceUpdate(userUUID uuid.UUID, deviceUUID uuid.UUID, deviceUpdate *models.DeviceUpdateRequest) (*models.Device, int, error) {
	user, err := s.userRepo.UserFindByUUID(userUUID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return nil, http.StatusForbidden, errors.New("only admin can update devices")
	}

	device, err := s.deviceRepo.DeviceFindByUUID(deviceUUID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("device not found")
	}

	if deviceUpdate.MacAddress == "" || deviceUpdate.Name == "" {
		return nil, http.StatusBadRequest, errors.New("name and mac address cannot be empty")
	}

	device.Name = deviceUpdate.Name

	updatedDevice, err := s.deviceRepo.DeviceUpdate(device)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &updatedDevice, http.StatusOK, nil
}
