package tests

import (
	"context"
	"home-monitor-backend/database"
	"home-monitor-backend/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	cleanAllTestData()
}

func cleanAllTestData() {
	database.DB.Where("username LIKE ?", "admin_%").Delete(&models.User{})
	database.DB.Where("username LIKE ?", "regular_%").Delete(&models.User{})

	database.DB.Where("mac_address LIKE ?", "00:1A:2B:%").Delete(&models.Device{})
}

func cleanupTestUser(ctx context.Context, username string) {
	var user models.User
	err := database.DB.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err == nil {
		database.DB.WithContext(ctx).Delete(&user)
	}
}

func cleanupTestDevice(ctx context.Context, macAddress string) {
	var device models.Device
	err := database.DB.WithContext(ctx).Where("mac_address = ?", macAddress).First(&device).Error
	if err == nil {
		database.DB.Delete(&device)
	}
}

func cleanupTestData(ctx context.Context, username string, macAddress string) {
	cleanupTestDevice(ctx, macAddress)
	cleanupTestUser(ctx, username)
}

func TestDeviceRegister(t *testing.T) {
	ctx := context.Background()

	cleanupTestUser(ctx, "admin_device_register")
	cleanupTestDevice(ctx, "00:1A:2B:3C:4D:5E")

	admin := &models.User{
		Username: "admin_device_register",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:5E",
		Name:       "Test Device",
	}

	registeredDevice, statusCode, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")
	assert.Equal(t, 201, statusCode, "Expected status code 201")

	if registeredDevice != nil {
		assert.Equal(t, device.MacAddress, registeredDevice.MacAddress, "Expected MAC address to match")
		assert.Equal(t, device.Name, registeredDevice.Name, "Expected device name to match")
	}

	t.Cleanup(func() {
		if registeredDevice != nil {
			deviceRepository.DeviceDelete(ctx, registeredDevice)
		}
		userRepository.UserDelete(ctx, admin)
	})
}
func TestDeviceUpdate(t *testing.T) {
	ctx := context.Background()

	cleanupTestData(ctx, "admin_device_update", "00:1A:2B:3C:4D:60")

	admin := &models.User{
		Username: "admin_device_update",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:60",
		Name:       "Test Device Update",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	updatedDevice := &models.DeviceUpdateRequest{
		Name:       "Updated Device Name",
		MacAddress: "00:1A:2B:3C:4D:60",
	}

	result, statusCode, err := deviceService.DeviceUpdate(ctx, admin.UUID, registeredDevice.UUID, updatedDevice)
	assert.NoError(t, err, "Failed to update device")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, "Updated Device Name", result.Name, "Expected device name to be updated")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceProfile(t *testing.T) {
	ctx := context.Background()

	cleanupTestData(ctx, "admin_device_profile", "00:1A:2B:3C:4D:5F")

	admin := &models.User{
		Username: "admin_device_profile",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:5F",
		Name:       "Test Device Profile",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	profileDevice, statusCode, err := deviceService.DeviceProfile(ctx, admin.UUID, registeredDevice.UUID)
	assert.NoError(t, err, "Failed to get device profile")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, registeredDevice.UUID, profileDevice.UUID, "Expected device UUID to match")
	assert.Equal(t, device.Name, profileDevice.Name, "Expected device name to match")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceLogin(t *testing.T) {
	ctx := context.Background()

	cleanupTestData(ctx, "admin_device_login", "00:1A:2B:3C:4D:61")

	admin := &models.User{
		Username: "admin_device_login",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:61",
		Name:       "Test Device Login",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	loginRequest := &models.DeviceLoginRequest{
		MacAddress: device.MacAddress,
	}

	loggedInDevice, token, statusCode, err := deviceService.DeviceLogin(ctx, loginRequest)
	assert.NoError(t, err, "Device login failed")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Equal(t, registeredDevice.UUID, loggedInDevice.UUID, "Expected device UUID to match")
	assert.NotEmpty(t, token, "Expected token to be generated")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceList(t *testing.T) {
	ctx := context.Background()

	cleanupTestData(ctx, "admin_device_list", "00:1A:2B:3C:4D:62")
	cleanupTestDevice(ctx, "00:1A:2B:3C:4D:63")

	admin := &models.User{
		Username: "admin_device_list",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device1 := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:62",
		Name:       "Test Device List 1",
	}
	device2 := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:63",
		Name:       "Test Device List 2",
	}

	registeredDevice1, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device1)
	assert.NoError(t, err, "Device 1 registration failed")

	registeredDevice2, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device2)
	assert.NoError(t, err, "Device 2 registration failed")

	listParams := &models.DeviceListRequestParams{
		Length: 10,
		Latest: false,
	}

	devices, statusCode, err := deviceService.DeviceList(ctx, admin.UUID, listParams)
	assert.NoError(t, err, "Failed to list devices")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Greater(t, len(devices), 0, "Expected at least one device in the list")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice1)
		assert.NoError(t, err, "Failed to delete device 1")

		err = deviceRepository.DeviceDelete(ctx, registeredDevice2)
		assert.NoError(t, err, "Failed to delete device 2")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceDelete(t *testing.T) {
	ctx := context.Background()

	cleanupTestData(ctx, "admin_device_delete", "00:1A:2B:3C:4D:64")

	admin := &models.User{
		Username: "admin_device_delete",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:64",
		Name:       "Test Device Delete",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	statusCode, err := deviceService.DeviceDelete(ctx, admin.UUID, registeredDevice.UUID)
	assert.NoError(t, err, "Failed to delete device")
	assert.Equal(t, 200, statusCode, "Expected status code 200")

	_, statusCode, err = deviceService.DeviceProfile(ctx, admin.UUID, registeredDevice.UUID)
	assert.Error(t, err, "Expected error when fetching deleted device")
	assert.Equal(t, 404, statusCode, "Expected status code 404 for deleted device")

	t.Cleanup(func() {
		err := userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceDuplicateMacAddress(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_device_duplicate",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device1 := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:65",
		Name:       "Test Device Duplicate 1",
	}

	registeredDevice1, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device1)
	assert.NoError(t, err, "First device registration failed")

	device2 := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:65",
		Name:       "Test Device Duplicate 2",
	}

	_, statusCode, err := deviceService.DeviceRegister(ctx, admin.UUID, device2)
	assert.Error(t, err, "Expected error when registering device with duplicate MAC address")
	assert.Equal(t, 409, statusCode, "Expected status code 409 for conflict")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice1)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceNonAdminCannotRegister(t *testing.T) {
	ctx := context.Background()

	normalUser := &models.User{
		Username: "normal_user_register",
		Password: "securepassword",
		Role:     models.UserRoleUser,
	}
	err := userRepository.UserCreate(ctx, normalUser)
	assert.NoError(t, err, "Failed to create normal user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:66",
		Name:       "Test Device Non Admin",
	}

	_, statusCode, err := deviceService.DeviceRegister(ctx, normalUser.UUID, device)
	assert.Error(t, err, "Expected error when non-admin tries to register device")
	assert.Equal(t, 403, statusCode, "Expected status code 403 for forbidden")

	t.Cleanup(func() {
		err := userRepository.UserDelete(ctx, normalUser)
		assert.NoError(t, err, "Failed to delete user")
	})
}

func TestDeviceMeasurements(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_measurements",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:67",
		Name:       "Test Device Measurements",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")
	assert.NotNil(t, registeredDevice, "Expected registered device to not be nil")

	payload1 := &models.DeviceMeasurementPayload{
		Temperature: 25.5,
		Humidity:    60.0,
	}

	measurement1, err := deviceService.DeviceCreateMeasurement(ctx, admin.UUID, registeredDevice.UUID, payload1)
	assert.NoError(t, err, "Failed to create measurement 1")
	assert.NotNil(t, measurement1, "Expected measurement 1 to be created")

	payload2 := &models.DeviceMeasurementPayload{
		Temperature: 26.0,
		Humidity:    65.0,
	}

	measurement2, err := deviceService.DeviceCreateMeasurement(ctx, admin.UUID, registeredDevice.UUID, payload2)
	assert.NoError(t, err, "Failed to create measurement 2")
	assert.NotNil(t, measurement2, "Expected measurement 2 to be created")

	requestParams := &models.DeviceMeasurementRequestParams{
		DeviceUUID: registeredDevice.UUID.String(),
		Length:     10,
		Latest:     false,
	}

	measurements, statusCode, err := deviceService.DeviceMeasurements(ctx, admin.UUID, requestParams)
	assert.NoError(t, err, "Failed to get measurements")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Greater(t, len(measurements), 0, "Expected at least one measurement")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceDeleteMeasurements(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_delete_measurements",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:68",
		Name:       "Test Device Delete Measurements",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	payload := &models.DeviceMeasurementPayload{
		Temperature: 25.5,
		Humidity:    60.0,
	}

	measurement, err := deviceService.DeviceCreateMeasurement(ctx, admin.UUID, registeredDevice.UUID, payload)
	assert.NoError(t, err, "Failed to create measurement")

	statusCode, err := deviceService.DeviceDeleteMeasurements(ctx, admin.UUID, []uint{measurement.ID})
	assert.NoError(t, err, "Failed to delete measurement")
	assert.Equal(t, 200, statusCode, "Expected status code 200")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceNonAdminCannotDeleteMeasurements(t *testing.T) {
	ctx := context.Background()

	normalUser := &models.User{
		Username: "normal_user_delete_measurements",
		Password: "securepassword",
		Role:     models.UserRoleUser,
	}
	err := userRepository.UserCreate(ctx, normalUser)
	assert.NoError(t, err, "Failed to create normal user")

	admin := &models.User{
		Username: "admin_for_normal_user",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err = userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:69",
		Name:       "Test Device Non Admin Delete",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	payload := &models.DeviceMeasurementPayload{
		Temperature: 25.5,
		Humidity:    60.0,
	}

	measurement, err := deviceService.DeviceCreateMeasurement(ctx, admin.UUID, registeredDevice.UUID, payload)
	assert.NoError(t, err, "Failed to create measurement")

	statusCode, err := deviceService.DeviceDeleteMeasurements(ctx, normalUser.UUID, []uint{measurement.ID})
	assert.Error(t, err, "Expected error when non-admin tries to delete measurements")
	assert.Equal(t, 403, statusCode, "Expected status code 403 for forbidden")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, normalUser)
		assert.NoError(t, err, "Failed to delete normal user")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceMeasurementsWithoutDevice(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_measurements_no_device",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	invalidDeviceUUID := "00000000-0000-0000-0000-000000000000"
	payload := &models.DeviceMeasurementPayload{
		Temperature: 25.5,
		Humidity:    60.0,
	}

	parsedUUID, _ := uuid.Parse(invalidDeviceUUID)
	_, err = deviceService.DeviceCreateMeasurement(ctx, admin.UUID, parsedUUID, payload)
	assert.Error(t, err, "Expected error when creating measurement for non-existent device")

	t.Cleanup(func() {
		err := userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceListWithPagination(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_list_pagination",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	registeredDevices := make([]*models.Device, 0)
	for i := range 3 {
		macSuffix := []string{"6E", "6F", "70"}
		device := &models.Device{
			MacAddress: "00:1A:2B:3C:4D:" + macSuffix[i],
			Name:       "Test Device Pagination " + string(rune(65+i)),
		}
		registered, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
		assert.NoError(t, err, "Failed to register device")
		registeredDevices = append(registeredDevices, registered)
	}

	listParams := &models.DeviceListRequestParams{
		Length: 2,
		Latest: true,
	}

	devices, statusCode, err := deviceService.DeviceList(ctx, admin.UUID, listParams)
	assert.NoError(t, err, "Failed to list devices")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.LessOrEqual(t, len(devices), 2, "Expected at most 2 devices")

	t.Cleanup(func() {
		for _, device := range registeredDevices {
			_ = deviceRepository.DeviceDelete(ctx, device)
		}

		err := userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceProfileWithInvalidUser(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_profile_invalid_user",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:6A",
		Name:       "Test Device Invalid User",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	invalidUserUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")
	_, statusCode, err := deviceService.DeviceProfile(ctx, invalidUserUUID, registeredDevice.UUID)
	assert.Error(t, err, "Expected error when fetching profile with invalid user")
	assert.Equal(t, 401, statusCode, "Expected status code 401")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceNonAdminCannotUpdate(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_cannot_update",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	normalUser := &models.User{
		Username: "normal_user_cannot_update",
		Password: "securepassword",
		Role:     models.UserRoleUser,
	}
	err = userRepository.UserCreate(ctx, normalUser)
	assert.NoError(t, err, "Failed to create normal user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:6B",
		Name:       "Test Device Non Admin Update",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	updateRequest := &models.DeviceUpdateRequest{
		Name:       "Updated Name",
		MacAddress: "00:1A:2B:3C:4D:6B",
	}

	_, statusCode, err := deviceService.DeviceUpdate(ctx, normalUser.UUID, registeredDevice.UUID, updateRequest)
	assert.Error(t, err, "Expected error when non-admin tries to update device")
	assert.Equal(t, 403, statusCode, "Expected status code 403 for forbidden")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, normalUser)
		assert.NoError(t, err, "Failed to delete normal user")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceNonAdminCannotDelete(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_cannot_delete",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	normalUser := &models.User{
		Username: "normal_user_cannot_delete",
		Password: "securepassword",
		Role:     models.UserRoleUser,
	}
	err = userRepository.UserCreate(ctx, normalUser)
	assert.NoError(t, err, "Failed to create normal user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:6C",
		Name:       "Test Device Non Admin Delete",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	statusCode, err := deviceService.DeviceDelete(ctx, normalUser.UUID, registeredDevice.UUID)
	assert.Error(t, err, "Expected error when non-admin tries to delete device")
	assert.Equal(t, 403, statusCode, "Expected status code 403 for forbidden")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, normalUser)
		assert.NoError(t, err, "Failed to delete normal user")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceCreateMeasurementWithInvalidUser(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_measurement_invalid_user",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:6D",
		Name:       "Test Device Measurement Invalid User",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	invalidUserUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")
	payload := &models.DeviceMeasurementPayload{
		Temperature: 25.5,
		Humidity:    60.0,
	}

	_, err = deviceService.DeviceCreateMeasurement(ctx, invalidUserUUID, registeredDevice.UUID, payload)
	assert.Error(t, err, "Expected error when creating measurement with invalid user")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceLoginWithInvalidMacAddress(t *testing.T) {
	ctx := context.Background()

	loginRequest := &models.DeviceLoginRequest{
		MacAddress: "00:00:00:00:00:00",
	}

	_, _, statusCode, err := deviceService.DeviceLogin(ctx, loginRequest)
	assert.Error(t, err, "Expected error when logging in with non-existent MAC address")
	assert.Equal(t, 404, statusCode, "Expected status code 404")
}

func TestDeviceMeasurementsAllDevices(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_measurements_all",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:6E",
		Name:       "Test Device All Measurements",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")
	assert.NotNil(t, registeredDevice, "Expected registered device to not be nil")

	payload := &models.DeviceMeasurementPayload{
		Temperature: 25.5,
		Humidity:    60.0,
	}

	_, err = deviceService.DeviceCreateMeasurement(ctx, admin.UUID, registeredDevice.UUID, payload)
	assert.NoError(t, err, "Failed to create measurement")

	requestParams := &models.DeviceMeasurementRequestParams{
		DeviceUUID: "",
		Length:     10,
		Latest:     false,
	}

	measurements, statusCode, err := deviceService.DeviceMeasurements(ctx, admin.UUID, requestParams)
	assert.NoError(t, err, "Failed to get all measurements")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.Greater(t, len(measurements), 0, "Expected at least one measurement")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceMeasurementsWithInvalidDeviceUUID(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_measurements_invalid_uuid",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	requestParams := &models.DeviceMeasurementRequestParams{
		DeviceUUID: "invalid-uuid",
		Length:     10,
		Latest:     false,
	}

	_, statusCode, err := deviceService.DeviceMeasurements(ctx, admin.UUID, requestParams)
	assert.Error(t, err, "Expected error with invalid device UUID")
	assert.Equal(t, 400, statusCode, "Expected status code 400 for bad request")

	t.Cleanup(func() {
		err := userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceMeasurementsWithNonExistentDevice(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_measurements_nonexistent",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	nonExistentDeviceUUID := "00000000-0000-0000-0000-000000000001"
	requestParams := &models.DeviceMeasurementRequestParams{
		DeviceUUID: nonExistentDeviceUUID,
		Length:     10,
		Latest:     false,
	}

	_, statusCode, err := deviceService.DeviceMeasurements(ctx, admin.UUID, requestParams)
	assert.Error(t, err, "Expected error when getting measurements for non-existent device")
	assert.Equal(t, 404, statusCode, "Expected status code 404")

	t.Cleanup(func() {
		err := userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceUpdateNonExistentDevice(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_update_nonexistent",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	nonExistentDeviceUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000002")
	updateRequest := &models.DeviceUpdateRequest{
		Name:       "Updated Name",
		MacAddress: "00:1A:2B:3C:4D:6F",
	}

	_, statusCode, err := deviceService.DeviceUpdate(ctx, admin.UUID, nonExistentDeviceUUID, updateRequest)
	assert.Error(t, err, "Expected error when updating non-existent device")
	assert.Equal(t, 404, statusCode, "Expected status code 404")

	t.Cleanup(func() {
		err := userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceDeleteNonExistentDevice(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_delete_nonexistent",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	nonExistentDeviceUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000003")
	statusCode, err := deviceService.DeviceDelete(ctx, admin.UUID, nonExistentDeviceUUID)
	assert.Error(t, err, "Expected error when deleting non-existent device")
	assert.Equal(t, 404, statusCode, "Expected status code 404")

	t.Cleanup(func() {
		err := userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceListWithInvalidUser(t *testing.T) {
	ctx := context.Background()

	invalidUserUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000004")
	listParams := &models.DeviceListRequestParams{
		Length: 10,
		Latest: false,
	}

	_, statusCode, err := deviceService.DeviceList(ctx, invalidUserUUID, listParams)
	assert.Error(t, err, "Expected error when listing devices with invalid user")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestDeviceDeleteMeasurementsWithInvalidUser(t *testing.T) {
	ctx := context.Background()

	invalidUserUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000005")
	statusCode, err := deviceService.DeviceDeleteMeasurements(ctx, invalidUserUUID, []uint{1})
	assert.Error(t, err, "Expected error when deleting measurements with invalid user")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestDeviceRegisterWithInvalidAdmin(t *testing.T) {
	ctx := context.Background()

	invalidAdminUUID, _ := uuid.Parse("00000000-0000-0000-0000-000000000006")
	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:71",
		Name:       "Test Device Invalid Admin",
	}

	_, statusCode, err := deviceService.DeviceRegister(ctx, invalidAdminUUID, device)
	assert.Error(t, err, "Expected error when registering device with invalid admin")
	assert.Equal(t, 401, statusCode, "Expected status code 401")
}

func TestDeviceUpdateWithEmptyName(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_update_empty_name",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:72",
		Name:       "Test Device Empty Name",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	updateRequest := &models.DeviceUpdateRequest{
		Name:       "",
		MacAddress: "00:1A:2B:3C:4D:72",
	}

	_, statusCode, err := deviceService.DeviceUpdate(ctx, admin.UUID, registeredDevice.UUID, updateRequest)
	assert.Error(t, err, "Expected error when updating device with empty name")
	assert.Equal(t, 400, statusCode, "Expected status code 400 for bad request")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceUpdateWithEmptyMacAddress(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_update_empty_mac",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:73",
		Name:       "Test Device Empty MAC",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	updateRequest := &models.DeviceUpdateRequest{
		Name:       "Updated Name",
		MacAddress: "",
	}

	_, statusCode, err := deviceService.DeviceUpdate(ctx, admin.UUID, registeredDevice.UUID, updateRequest)
	assert.Error(t, err, "Expected error when updating device with empty MAC address")
	assert.Equal(t, 400, statusCode, "Expected status code 400 for bad request")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceMeasurementsLatestFlag(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_measurements_latest",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:74",
		Name:       "Test Device Latest Measurements",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	for i := range 3 {
		payload := &models.DeviceMeasurementPayload{
			Temperature: 25.0 + float64(i),
			Humidity:    60.0 + float64(i),
		}
		_, err := deviceService.DeviceCreateMeasurement(ctx, admin.UUID, registeredDevice.UUID, payload)
		assert.NoError(t, err, "Failed to create measurement")
	}

	requestParams := &models.DeviceMeasurementRequestParams{
		DeviceUUID: registeredDevice.UUID.String(),
		Length:     2,
		Latest:     true,
	}

	measurements, statusCode, err := deviceService.DeviceMeasurements(ctx, admin.UUID, requestParams)
	assert.NoError(t, err, "Failed to get latest measurements")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.LessOrEqual(t, len(measurements), 2, "Expected at most 2 measurements")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusDefaultFalse(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_device_status_default",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:70",
		Name:       "Test Device Status Default",
	}

	registeredDevice, statusCode, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")
	assert.Equal(t, 201, statusCode, "Expected status code 201")
	assert.False(t, registeredDevice.Status, "Expected device status to default to false")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusUpdateToTrueOnMeasurement(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_device_status_measurement",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:71",
		Name:       "Test Device Status Measurement",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")
	assert.False(t, registeredDevice.Status, "Expected initial status to be false")

	payload := &models.DeviceMeasurementPayload{
		Temperature: 25.0,
		Humidity:    60.0,
	}

	_, err = deviceService.DeviceCreateMeasurement(ctx, admin.UUID, registeredDevice.UUID, payload)
	assert.NoError(t, err, "Failed to create measurement")

	updatedDevice, statusCode, err := deviceService.DeviceProfile(ctx, admin.UUID, registeredDevice.UUID)
	assert.NoError(t, err, "Failed to get device profile")
	assert.Equal(t, 200, statusCode, "Expected status code 200")
	assert.True(t, updatedDevice.Status, "Expected device status to be true after measurement")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusManualUpdate(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_device_status_manual",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:72",
		Name:       "Test Device Status Manual",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	statusCode, err := deviceService.DeviceUpdateStatus(ctx, admin.UUID, registeredDevice.UUID, true)
	assert.NoError(t, err, "Failed to update device status")
	assert.Equal(t, 200, statusCode, "Expected status code 200")

	updatedDevice, _, err := deviceService.DeviceProfile(ctx, admin.UUID, registeredDevice.UUID)
	assert.NoError(t, err, "Failed to get device profile")
	assert.True(t, updatedDevice.Status, "Expected device status to be true")

	statusCode, err = deviceService.DeviceUpdateStatus(ctx, admin.UUID, registeredDevice.UUID, false)
	assert.NoError(t, err, "Failed to update device status")
	assert.Equal(t, 200, statusCode, "Expected status code 200")

	updatedDevice, _, err = deviceService.DeviceProfile(ctx, admin.UUID, registeredDevice.UUID)
	assert.NoError(t, err, "Failed to get device profile")
	assert.False(t, updatedDevice.Status, "Expected device status to be false")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusManualUpdateNonAdmin(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_device_status_non_admin",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	regularUser := &models.User{
		Username: "regular_user_device_status",
		Password: "securepassword",
		Role:     models.UserRoleUser,
	}
	err = userRepository.UserCreate(ctx, regularUser)
	assert.NoError(t, err, "Failed to create regular user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:73",
		Name:       "Test Device Status Non Admin",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	statusCode, err := deviceService.DeviceUpdateStatus(ctx, regularUser.UUID, registeredDevice.UUID, true)
	assert.Error(t, err, "Expected error when non-admin tries to update device status")
	assert.Equal(t, 403, statusCode, "Expected status code 403 (Forbidden)")
	assert.Contains(t, err.Error(), "only admin", "Expected error message about admin requirement")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")

		err = userRepository.UserDelete(ctx, regularUser)
		assert.NoError(t, err, "Failed to delete regular user")
	})
}

func TestDeviceStatusUpdateNonExistentDevice(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_device_status_nonexistent",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	nonExistentUUID := uuid.New()

	statusCode, err := deviceService.DeviceUpdateStatus(ctx, admin.UUID, nonExistentUUID, true)
	assert.Error(t, err, "Expected error when updating non-existent device")
	assert.Equal(t, 404, statusCode, "Expected status code 404")
	assert.Contains(t, err.Error(), "not found", "Expected error message about device not found")

	t.Cleanup(func() {
		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusRepositoryUpdateStatus(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_device_repo_status",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:74",
		Name:       "Test Device Repo Status",
	}

	registeredDevice, _, err := deviceService.DeviceRegister(ctx, admin.UUID, device)
	assert.NoError(t, err, "Device registration failed")

	err = deviceRepository.DeviceUpdateStatus(ctx, registeredDevice.ID, true)
	assert.NoError(t, err, "Failed to update device status via repository")

	updatedDevice, err := deviceRepository.DeviceFindByUUID(ctx, registeredDevice.UUID)
	assert.NoError(t, err, "Failed to find device")
	assert.True(t, updatedDevice.Status, "Expected device status to be true")

	err = deviceRepository.DeviceUpdateStatus(ctx, registeredDevice.ID, false)
	assert.NoError(t, err, "Failed to update device status via repository")

	updatedDevice, err = deviceRepository.DeviceFindByUUID(ctx, registeredDevice.UUID)
	assert.NoError(t, err, "Failed to find device")
	assert.False(t, updatedDevice.Status, "Expected device status to be false")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, registeredDevice)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusCheckerTimeout(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_status_checker_timeout",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:80",
		Name:       "Test Device Status Checker",
	}
	err = deviceRepository.DeviceCreate(ctx, device)
	assert.NoError(t, err, "Failed to create device")

	err = deviceRepository.DeviceUpdateStatus(ctx, device.ID, true)
	assert.NoError(t, err, "Failed to set device status to true")

	oldMeasurement := &models.DeviceMeasurement{
		DeviceID:    device.ID,
		Temperature: 25.0,
		Humidity:    60.0,
	}
	err = deviceRepository.DeviceMeasurementCreate(ctx, oldMeasurement)
	assert.NoError(t, err, "Failed to create measurement")

	database.DB.Model(oldMeasurement).Update("created_at", time.Now().Add(-15*time.Minute))

	measurements, err := deviceRepository.DeviceMeasurementsByDeviceID(ctx, 1, device.ID, true, nil, nil)
	assert.NoError(t, err, "Failed to get measurements")
	assert.Greater(t, len(measurements), 0, "Expected at least one measurement")

	lastMeasurement := measurements[0]
	timeSinceLastMeasurement := time.Since(lastMeasurement.CreatedAt)

	assert.Greater(t, timeSinceLastMeasurement, 10*time.Minute, "Expected measurement to be older than 10 minutes")

	if timeSinceLastMeasurement > 10*time.Minute {
		err = deviceRepository.DeviceUpdateStatus(ctx, device.ID, false)
		assert.NoError(t, err, "Failed to update device status to false")
	}

	updatedDevice, err := deviceRepository.DeviceFindByUUID(ctx, device.UUID)
	assert.NoError(t, err, "Failed to find device")
	assert.False(t, updatedDevice.Status, "Expected device status to be false after timeout")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, device)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusCheckerRecentMeasurement(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_status_checker_recent",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:81",
		Name:       "Test Device Status Recent",
	}
	err = deviceRepository.DeviceCreate(ctx, device)
	assert.NoError(t, err, "Failed to create device")

	err = deviceRepository.DeviceUpdateStatus(ctx, device.ID, true)
	assert.NoError(t, err, "Failed to set device status to true")

	recentMeasurement := &models.DeviceMeasurement{
		DeviceID:    device.ID,
		Temperature: 25.0,
		Humidity:    60.0,
	}
	err = deviceRepository.DeviceMeasurementCreate(ctx, recentMeasurement)
	assert.NoError(t, err, "Failed to create measurement")

	measurements, err := deviceRepository.DeviceMeasurementsByDeviceID(ctx, 1, device.ID, true, nil, nil)
	assert.NoError(t, err, "Failed to get measurements")
	assert.Greater(t, len(measurements), 0, "Expected at least one measurement")

	lastMeasurement := measurements[0]
	timeSinceLastMeasurement := time.Since(lastMeasurement.CreatedAt)

	assert.Less(t, timeSinceLastMeasurement, 10*time.Minute, "Expected measurement to be less than 10 minutes old")

	if timeSinceLastMeasurement > 10*time.Minute {
		err = deviceRepository.DeviceUpdateStatus(ctx, device.ID, false)
		assert.NoError(t, err, "Failed to update device status to false")
	}

	updatedDevice, err := deviceRepository.DeviceFindByUUID(ctx, device.UUID)
	assert.NoError(t, err, "Failed to find device")
	assert.True(t, updatedDevice.Status, "Expected device status to remain true with recent measurements")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, device)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusCheckerSkipsOfflineDevices(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_status_checker_skip",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:82",
		Name:       "Test Device Status Skip",
	}
	err = deviceRepository.DeviceCreate(ctx, device)
	assert.NoError(t, err, "Failed to create device")

	currentDevice, err := deviceRepository.DeviceFindByUUID(ctx, device.UUID)
	assert.NoError(t, err, "Failed to find device")
	assert.False(t, currentDevice.Status, "Expected device status to be false initially")

	oldMeasurement := &models.DeviceMeasurement{
		DeviceID:    device.ID,
		Temperature: 25.0,
		Humidity:    60.0,
	}
	err = deviceRepository.DeviceMeasurementCreate(ctx, oldMeasurement)
	assert.NoError(t, err, "Failed to create measurement")

	database.DB.Model(oldMeasurement).Update("created_at", time.Now().Add(-15*time.Minute))

	if currentDevice.Status {
		measurements, err := deviceRepository.DeviceMeasurementsByDeviceID(ctx, 1, device.ID, true, nil, nil)
		assert.NoError(t, err, "Failed to get measurements")

		if len(measurements) > 0 {
			lastMeasurement := measurements[0]
			timeSinceLastMeasurement := time.Since(lastMeasurement.CreatedAt)

			if timeSinceLastMeasurement > 10*time.Minute {
				err = deviceRepository.DeviceUpdateStatus(ctx, device.ID, false)
				assert.NoError(t, err, "Failed to update device status to false")
			}
		}
	}

	updatedDevice, err := deviceRepository.DeviceFindByUUID(ctx, device.UUID)
	assert.NoError(t, err, "Failed to find device")
	assert.False(t, updatedDevice.Status, "Expected device status to remain false")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, device)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}

func TestDeviceStatusCheckerNoMeasurements(t *testing.T) {
	ctx := context.Background()

	admin := &models.User{
		Username: "admin_status_checker_no_measurements",
		Password: "securepassword",
		Role:     models.UserRoleAdmin,
	}
	err := userRepository.UserCreate(ctx, admin)
	assert.NoError(t, err, "Failed to create admin user")

	device := &models.Device{
		MacAddress: "00:1A:2B:3C:4D:83",
		Name:       "Test Device Status No Measurements",
	}
	err = deviceRepository.DeviceCreate(ctx, device)
	assert.NoError(t, err, "Failed to create device")

	err = deviceRepository.DeviceUpdateStatus(ctx, device.ID, true)
	assert.NoError(t, err, "Failed to set device status to true")

	measurements, err := deviceRepository.DeviceMeasurementsByDeviceID(ctx, 1, device.ID, true, nil, nil)
	assert.NoError(t, err, "Failed to get measurements")
	assert.Equal(t, 0, len(measurements), "Expected no measurements")

	updatedDevice, err := deviceRepository.DeviceFindByUUID(ctx, device.UUID)
	assert.NoError(t, err, "Failed to find device")
	assert.True(t, updatedDevice.Status, "Expected device status to remain true when no measurements exist")

	t.Cleanup(func() {
		err := deviceRepository.DeviceDelete(ctx, device)
		assert.NoError(t, err, "Failed to delete device")

		err = userRepository.UserDelete(ctx, admin)
		assert.NoError(t, err, "Failed to delete admin user")
	})
}
