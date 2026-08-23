package services

import (
	"context"
	"errors"
	"home-monitor-backend/models"
	"home-monitor-backend/repositories"
	"home-monitor-backend/utils"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type DeviceService interface {
	DeviceRegister(ctx context.Context, userUUID uuid.UUID, device *models.Device) (*models.Device, int, error)
	DeviceLogin(ctx context.Context, deviceLoginRequest *models.DeviceLoginRequest) (*models.Device, string, int, error)
	DeviceDelete(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID) (int, error)
	DeviceList(ctx context.Context, userUUID uuid.UUID, requestParams *models.DeviceListRequestParams) ([]models.Device, int, error)
	DeviceProfile(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID) (*models.Device, int, error)
	DeviceUpdate(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID, deviceUpdate *models.DeviceUpdateRequest) (*models.Device, int, error)
	DeviceUpdateStatus(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID, status bool) (int, error)
	DeviceMeasurements(ctx context.Context, userUUID uuid.UUID, requestParams *models.DeviceMeasurementRequestParams) ([]models.DeviceMeasurement, int, error)
	DeviceCreateMeasurement(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID, payload *models.DeviceMeasurementPayload) (*models.DeviceMeasurement, error)
	DeviceDeleteMeasurements(ctx context.Context, userUUID uuid.UUID, measurementIDs []uint) (int, error)
}

type deviceService struct {
	deviceRepo repositories.DeviceRepository
	userRepo   repositories.UserRepository
}

func NewDeviceService(deviceRepo repositories.DeviceRepository, userRepo repositories.UserRepository) DeviceService {
	return &deviceService{deviceRepo: deviceRepo, userRepo: userRepo}
}

func (s *deviceService) DeviceRegister(ctx context.Context, userUUID uuid.UUID, device *models.Device) (*models.Device, int, error) {
	user, err := s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return nil, http.StatusForbidden, errors.New("only admin can create new devices")
	}

	existingDevice, _ := s.deviceRepo.DeviceFindByMacAddress(ctx, device.MacAddress)
	if existingDevice != nil {
		return nil, http.StatusConflict, errors.New("device with this MAC address already exists")
	}

	existingDevice, _ = s.deviceRepo.DeviceFindByName(ctx, device.Name)
	if existingDevice != nil {
		return nil, http.StatusConflict, errors.New("device with this name already exists")
	}

	if err := s.deviceRepo.DeviceCreate(ctx, device); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return device, http.StatusCreated, nil
}

func (s *deviceService) DeviceLogin(ctx context.Context, deviceLoginRequest *models.DeviceLoginRequest) (*models.Device, string, int, error) {
	device, err := s.deviceRepo.DeviceFindByMacAddress(ctx, deviceLoginRequest.MacAddress)
	if err != nil {
		return nil, "", http.StatusNotFound, errors.New("device not found")
	}

	token, err := utils.GenerateDeviceJWT(device.UUID)
	if err != nil {
		return nil, "", http.StatusInternalServerError, errors.New("failed to generate token")
	}

	return device, token, http.StatusOK, nil
}

func (s *deviceService) DeviceDelete(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID) (int, error) {
	user, err := s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return http.StatusUnauthorized, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return http.StatusForbidden, errors.New("only admin can delete devices")
	}

	device, err := s.deviceRepo.DeviceFindByUUID(ctx, deviceUUID)
	if err != nil {
		return http.StatusNotFound, errors.New("device not found")
	}

	measurements, err := s.deviceRepo.DeviceMeasurementsByDeviceID(ctx, 1, device.ID, false, nil, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if len(measurements) > 0 {
		return http.StatusForbidden, errors.New("cannot delete device with existing measurements")
	}

	if err := s.deviceRepo.DeviceDelete(ctx, device); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (s *deviceService) DeviceList(ctx context.Context, userUUID uuid.UUID, requestParams *models.DeviceListRequestParams) ([]models.Device, int, error) {
	_, err := s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("user not found")
	}

	devices, err := s.deviceRepo.DeviceList(ctx, requestParams.Length, requestParams.Latest)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return devices, http.StatusOK, nil
}

func (s *deviceService) DeviceProfile(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID) (*models.Device, int, error) {
	_, err := s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("user not found")
	}

	device, err := s.deviceRepo.DeviceFindByUUID(ctx, deviceUUID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("device not found")
	}
	return device, http.StatusOK, nil
}

func (s *deviceService) DeviceUpdate(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID, deviceUpdate *models.DeviceUpdateRequest) (*models.Device, int, error) {
	user, err := s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return nil, http.StatusForbidden, errors.New("only admin can update devices")
	}

	device, err := s.deviceRepo.DeviceFindByUUID(ctx, deviceUUID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("device not found")
	}

	if deviceUpdate.MacAddress == "" || deviceUpdate.Name == "" {
		return nil, http.StatusBadRequest, errors.New("name and mac address cannot be empty")
	}

	device.Name = deviceUpdate.Name

	updatedDevice, err := s.deviceRepo.DeviceUpdate(ctx, device)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &updatedDevice, http.StatusOK, nil
}

func (s *deviceService) DeviceMeasurements(ctx context.Context, userUUID uuid.UUID, requestParams *models.DeviceMeasurementRequestParams) ([]models.DeviceMeasurement, int, error) {
	var measurements []models.DeviceMeasurement
	var err error

	_, err = s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return nil, http.StatusUnauthorized, errors.New("user not found")
	}

	startTime := requestParams.StartTime
	endTime := requestParams.EndTime

	var startTimePtr *string
	var endTimePtr *string

	if startTime != "" {
		startTimePtr = &startTime
	} else {
		start := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
		startTimePtr = &start
	}

	if endTime != "" {
		endTimePtr = &endTime
	} else {
		end := time.Now().Format(time.RFC3339)
		endTimePtr = &end
	}

	if requestParams.DeviceUUID != "" {
		deviceUUID, parseErr := uuid.Parse(requestParams.DeviceUUID)
		if parseErr != nil {
			return nil, http.StatusBadRequest, errors.New("invalid device UUID")
		}

		device, err := s.deviceRepo.DeviceFindByUUID(ctx, deviceUUID)
		if err != nil {
			return nil, http.StatusNotFound, errors.New("device not found")
		}

		measurements, err = s.deviceRepo.DeviceMeasurementsByDeviceID(ctx, requestParams.Length, device.ID, requestParams.Latest, startTimePtr, endTimePtr)

		if len(measurements) == 0 {
			measurements, err = s.deviceRepo.DeviceMeasurementsByDeviceID(ctx, requestParams.Length, device.ID, requestParams.Latest, nil, nil)
		}

		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	} else {
		measurements, err = s.deviceRepo.DeviceMeasurements(ctx, requestParams.Length, requestParams.Latest, startTimePtr, endTimePtr)

		if len(measurements) == 0 {
			measurements, err = s.deviceRepo.DeviceMeasurements(ctx, requestParams.Length, requestParams.Latest, nil, nil)
		}

		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}

	return measurements, http.StatusOK, nil
}

func (s *deviceService) DeviceCreateMeasurement(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID, payload *models.DeviceMeasurementPayload) (*models.DeviceMeasurement, error) {
	_, err := s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	device, err := s.deviceRepo.DeviceFindByUUID(ctx, deviceUUID)
	if err != nil {
		return nil, errors.New("device not found")
	}

	measurement := &models.DeviceMeasurement{
		DeviceID:    device.ID,
		Temperature: payload.Temperature,
		Humidity:    payload.Humidity,
	}

	if err := s.deviceRepo.DeviceMeasurementCreate(ctx, measurement); err != nil {
		return nil, err
	}

	if err := s.deviceRepo.DeviceUpdateStatus(ctx, device.ID, true); err != nil {
		return nil, err
	}

	return measurement, nil
}

func (s *deviceService) DeviceUpdateStatus(ctx context.Context, userUUID uuid.UUID, deviceUUID uuid.UUID, status bool) (int, error) {
	user, err := s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return http.StatusUnauthorized, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return http.StatusForbidden, errors.New("only admin can update device status")
	}

	device, err := s.deviceRepo.DeviceFindByUUID(ctx, deviceUUID)
	if err != nil {
		return http.StatusNotFound, errors.New("device not found")
	}

	if err := s.deviceRepo.DeviceUpdateStatus(ctx, device.ID, status); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (s *deviceService) DeviceDeleteMeasurements(ctx context.Context, userUUID uuid.UUID, measurementIDs []uint) (int, error) {
	user, err := s.userRepo.UserFindByUUID(ctx, userUUID)
	if err != nil {
		return http.StatusUnauthorized, errors.New("user not found")
	}

	if user.Role != models.UserRoleAdmin {
		return http.StatusForbidden, errors.New("only admin can delete measurements")
	}

	if err := s.deviceRepo.DeviceMeasurementDelete(ctx, measurementIDs); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
