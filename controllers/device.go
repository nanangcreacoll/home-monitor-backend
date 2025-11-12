package controllers

import (
	"home-monitor-backend/models"
	"home-monitor-backend/services"
	"net/http"
	"strconv"

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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	device := &models.Device{
		MacAddress: input.MacAddress,
		Name:       input.Name,
	}

	createdDevice, statusCode, err := ctrl.deviceService.DeviceRegister(device, userUUID.(uuid.UUID))
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(statusCode, models.DeviceResponse{
		UUID:          createdDevice.UUID,
		MacAddress:    createdDevice.MacAddress,
		Name:          createdDevice.Name,
		UserCreatedID: createdDevice.UserCreatedID,
		CreatedAt:     createdDevice.CreatedAt,
		UpdatedAt:     createdDevice.UpdatedAt,
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

	var input models.DeviceDeleteRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	statusCode, err := ctrl.deviceService.DeviceDelete(input.UUID, userUUID.(uuid.UUID))
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
// @Router /devices [get]
func (ctrl *DeviceController) DeviceList(c *gin.Context) {
	userUUID, exists := c.Get("userUUID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	lengthParam := c.DefaultQuery("length", "0")
	latestParam := c.DefaultQuery("latest", "false")

	length := 0
	if l, err := strconv.Atoi(lengthParam); err == nil && l > 0 {
		length = l
	}

	latest := false
	if latestParam == "true" {
		latest = true
	}

	devices, statusCode, err := ctrl.deviceService.DeviceList(userUUID.(uuid.UUID), length, latest)
	if err != nil {
		c.JSON(statusCode, models.ErrorResponse{Error: err.Error()})
		return
	}

	var deviceResponses []models.DeviceResponse
	for _, device := range devices {
		deviceResponses = append(deviceResponses, models.DeviceResponse{
			UUID:          device.UUID,
			MacAddress:    device.MacAddress,
			Name:          device.Name,
			UserCreatedID: device.UserCreatedID,
			CreatedAt:     device.CreatedAt,
			UpdatedAt:     device.UpdatedAt,
		})
	}

	c.JSON(statusCode, models.DeviceListResponse{Devices: deviceResponses})
}
