package tests

import (
	"home-monitor-backend/database"
	"home-monitor-backend/repositories"
	"home-monitor-backend/services"
	"home-monitor-backend/utils"
	"log"

	"github.com/joho/godotenv"
)

var (
	userRepository   repositories.UserRepository
	deviceRepository repositories.DeviceRepository
	userService      services.UserService
	deviceService    services.DeviceService
)

func init() {
	envPath := utils.FindDotEnv(3)
	if envPath == "" {
		log.Println("Could not find .env file")
	}

	if envPath != "" {
		err := godotenv.Load(envPath)
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	database.ConnectDB()
	database.Migrations()

	userRepository = repositories.NewUserRepository()
	deviceRepository = repositories.NewDeviceRepository()
	userService = services.NewUserService(userRepository, deviceRepository)
	deviceService = services.NewDeviceService(deviceRepository, userRepository)
}
