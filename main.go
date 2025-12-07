package main

import (
	"home-monitor-backend/controllers"
	"home-monitor-backend/database"
	"home-monitor-backend/docs"
	"home-monitor-backend/pkg"
	"home-monitor-backend/repositories"
	"home-monitor-backend/routes"
	"home-monitor-backend/services"
	"home-monitor-backend/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Home Monitor API
// @version 1.0
// @description API for Home Monitoring System

// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	envPath := utils.FindDotEnv(3)
	if envPath == "" {
		log.Fatal("Could not find .env file")
	}

	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.ConnectDB()
	database.Migrations()

	if os.Args[len(os.Args)-1] == "--seed" || os.Args[len(os.Args)-1] == "-s" {
		err := database.Seed()
		if err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		log.Println("Seeding completed successfully")
		return
	}

	userRepo := repositories.NewUserRepository()
	deviceRepo := repositories.NewDeviceRepository()
	userService := services.NewUserService(userRepo, deviceRepo)
	deviceService := services.NewDeviceService(deviceRepo, userRepo)
	userController := controllers.NewUserController(userService)
	deviceController := controllers.NewDeviceController(deviceService)

	if os.Getenv("GIN_MODE") != "release" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	routes.RootRoute(r)
	routes.UserRoutes(r, userController)
	routes.DeviceRoutes(r, deviceController)

	docs.SwaggerInfo.BasePath = "/api"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	err = pkg.MqttInit()
	if err != nil {
		log.Fatalf("Failed to initialize MQTT client: %v", err)
	}

	go pkg.DeviceInit(deviceRepo)

	r.Run(":8080")
}
