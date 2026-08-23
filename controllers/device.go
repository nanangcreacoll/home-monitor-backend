package controllers

import (
	"home-monitor-backend/models"
	"home-monitor-backend/services"
	"home-monitor-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DeviceController struct {
	deviceService services.DeviceService
}

func NewDeviceController(deviceService services.DeviceService) *DeviceController {
	return &DeviceController{deviceService: deviceService}
}

// DeviceCreate godoc
// @Summary Create new device
// @Description Create a new device
// @Tags devices
// @Accept json
// @Produce json
// @Param request body models.DeviceCreateRequest true "Device create request"
// @Success 201 {object} models.DeviceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device [post]
func (ctrl *DeviceController) DeviceRegister(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var input models.DeviceCreateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	device := &models.Device{
		MacAddress: input.MacAddress,
		Name:       input.Name,
	}

	createdDevice, statusCode, err := ctrl.deviceService.DeviceRegister(c, userUUID.(uuid.UUID), device)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.DeviceResponse{
		UUID:       createdDevice.UUID,
		MacAddress: createdDevice.MacAddress,
		Name:       createdDevice.Name,
		CreatedAt:  createdDevice.CreatedAt,
		UpdatedAt:  createdDevice.UpdatedAt,
		Status:     createdDevice.Status,
	})
}

// DeviceLogin godoc
// @Summary Device login
// @Description Authenticate device and return JWT token
// @Tags devices
// @Accept json
// @Produce json
// @Param request body models.DeviceLoginRequest true "Device login request"
// @Success 200 {object} models.DeviceLoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /device/login [post]
func (ctrl *DeviceController) DeviceLogin(c *gin.Context) {
	var input models.DeviceLoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	device, token, statusCode, err := ctrl.deviceService.DeviceLogin(c, &input)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.DeviceLoginResponse{
		UUID:  device.UUID,
		Token: token,
	})
}

// DeviceDelete godoc
// @Summary Delete a device
// @Description Delete a device by UUID
// @Tags devices
// @Accept json
// @Produce json
// @Param uuid path string true "Device UUID"
// @Success 200 {object} models.ResponseMessage
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device/{uuid} [delete]
func (ctrl *DeviceController) DeviceDelete(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	deviceUUID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device UUID"})
		return
	}

	deviceDeleteRequest := models.DeviceDeleteRequest{
		UUID: deviceUUID,
	}

	statusCode, err := ctrl.deviceService.DeviceDelete(c, userUUID.(uuid.UUID), deviceDeleteRequest.UUID)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.ResponseMessage{Message: "Device deleted successfully"})
}

// DeviceList godoc
// @Summary List devices
// @Description List all devices with optional length and latest parameters
// @Tags devices
// @Produce json
// @Param length query int false "Number of devices to return"
// @Param latest query bool false "Return the latest devices first"
// @Success 200 {object} models.DeviceListResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device/list [get]
func (ctrl *DeviceController) DeviceList(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var requestParams models.DeviceListRequestParams
	if err := c.ShouldBindQuery(&requestParams); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	devices, statusCode, err := ctrl.deviceService.DeviceList(c, userUUID.(uuid.UUID), &requestParams)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	var deviceResponses []models.DeviceResponse
	for _, device := range devices {
		deviceResponses = append(deviceResponses, models.DeviceResponse{
			UUID:       device.UUID,
			MacAddress: device.MacAddress,
			Name:       device.Name,
			CreatedAt:  device.CreatedAt,
			UpdatedAt:  device.UpdatedAt,
			Status:     device.Status,
		})
	}

	c.JSON(statusCode, models.DeviceListResponse{Devices: deviceResponses})
}

// DeviceProfile godoc
// @Summary Get device profile
// @Description Get the profile of a device by UUID
// @Tags devices
// @Produce json
// @Param uuid path string true "Device UUID"
// @Success 200 {object} models.DeviceResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device/{uuid} [get]
func (ctrl *DeviceController) DeviceProfile(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	deviceUUIDParam := c.Param("uuid")
	deviceUUID, err := uuid.Parse(deviceUUIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device UUID"})
		return
	}

	device, statusCode, err := ctrl.deviceService.DeviceProfile(c, userUUID.(uuid.UUID), deviceUUID)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.DeviceResponse{
		UUID:       device.UUID,
		MacAddress: device.MacAddress,
		Name:       device.Name,
		CreatedAt:  device.CreatedAt,
		UpdatedAt:  device.UpdatedAt,
		Status:     device.Status,
	})
}

// DeviceUpdate godoc
// @Summary Update device
// @Description Update the details of a device by UUID
// @Tags devices
// @Accept json
// @Produce json
// @Param uuid path string true "Device UUID"
// @Param request body models.DeviceUpdateRequest true "Device update request"
// @Success 200 {object} models.DeviceResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device/{uuid} [put]
func (ctrl *DeviceController) DeviceUpdate(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	deviceUUIDParam := c.Param("uuid")
	deviceUUID, err := uuid.Parse(deviceUUIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device UUID"})
		return
	}

	var input models.DeviceUpdateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	updatedDevice, statusCode, err := ctrl.deviceService.DeviceUpdate(c, userUUID.(uuid.UUID), deviceUUID, &input)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.DeviceResponse{
		UUID:       updatedDevice.UUID,
		MacAddress: updatedDevice.MacAddress,
		Name:       updatedDevice.Name,
		CreatedAt:  updatedDevice.CreatedAt,
		UpdatedAt:  updatedDevice.UpdatedAt,
		Status:     updatedDevice.Status,
	})
}

// DeviceUpdateStatus godoc
// @Summary Update device status
// @Description Update device status (online/offline) - Admin only
// @Tags devices
// @Accept json
// @Produce json
// @Param uuid path string true "Device UUID"
// @Param request body models.DeviceUpdateStatusRequest true "Device status update request"
// @Success 200 {object} models.ResponseMessage
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device/{uuid}/status [patch]
func (ctrl *DeviceController) DeviceUpdateStatus(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	deviceUUIDParam := c.Param("uuid")
	deviceUUID, err := uuid.Parse(deviceUUIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device UUID"})
		return
	}

	var input models.DeviceUpdateStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	statusCode, err := ctrl.deviceService.DeviceUpdateStatus(c, userUUID.(uuid.UUID), deviceUUID, input.Status)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.ResponseMessage{Message: "Device status updated successfully"})
}

// DeviceMeasurements godoc
// @Summary Get device measurements
// @Description Get measurements for a device with optional length and latest parameters
// @Tags devices
// @Produce json
// @Param device_uuid query string false "Device UUID"
// @Param length query int false "Number of measurements to return"
// @Param latest query bool false "Return the latest measurements first"
// @Success 200 {object} models.DeviceMeasurementListResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device/measurements [get]
func (ctrl *DeviceController) DeviceMeasurements(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var requestParams models.DeviceMeasurementRequestParams
	if err := c.ShouldBindQuery(&requestParams); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	measurements, statusCode, err := ctrl.deviceService.DeviceMeasurements(c, userUUID.(uuid.UUID), &requestParams)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(
		statusCode,
		models.DeviceMeasurementListResponse{Measurements: measurements},
	)
}

// DeviceCreateMeasurement godoc
// @Summary Create device measurement
// @Description Create a new measurement for a device by UUID
// @Tags devices
// @Accept json
// @Produce json
// @Param uuid path string true "Device UUID"
// @Param request body models.DeviceMeasurementPayload true "Device measurement payload"
// @Success 201 {object} models.DeviceMeasurement
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device/{uuid}/measurements [post]
func (ctrl *DeviceController) DeviceCreateMeasurement(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	deviceUUIDParam := c.Param("uuid")
	deviceUUID, err := uuid.Parse(deviceUUIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid device UUID"})
		return
	}

	var input models.DeviceMeasurementPayload
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	createdMeasurement, err := ctrl.deviceService.DeviceCreateMeasurement(c, userUUID.(uuid.UUID), deviceUUID, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdMeasurement)
}

// DeviceDeleteMeasurements godoc
// @Summary Delete device measurements
// @Description Delete measurements for a device by measurement IDs
// @Tags devices
// @Accept json
// @Produce json
// @Param request body models.DeviceMeasurementDeleteRequest true "Device measurement delete request"
// @Success 200 {object} models.ResponseMessage
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /device/measurements [delete]
func (ctrl *DeviceController) DeviceDeleteMeasurements(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	var input models.DeviceMeasurementDeleteRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		errors := utils.ValidationError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errors})
		return
	}

	statusCode, err := ctrl.deviceService.DeviceDeleteMeasurements(c, userUUID.(uuid.UUID), input.IDs)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.ResponseMessage{Message: "Device measurements deleted successfully"})
}
