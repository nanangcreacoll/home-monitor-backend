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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Home Monitor API
// @version 1.0
// @description API for Home Monitoring System

type Args string

const (
	GenerateSecretJWT      Args = "--generate-secret"
	GenerateSecretJWTShort Args = "-gs"
	MigrateDown            Args = "--migrate-down"
	MigrateDownShort       Args = "-md"
	Seed                   Args = "--seed"
	SeedShort              Args = "-s"
)

// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if len(os.Args) > 2 {
		log.Fatal("Too many arguments provided")
		return
	} else if !strings.HasPrefix(os.Args[len(os.Args)-1], "-") && len(os.Args) == 2 {
		log.Fatal("Invalid argument provided")
		return
	}

	envPath := utils.FindDotEnv(3)
	if envPath == "" {
		log.Fatal("Could not find .env file")
	}

	if os.Args[len(os.Args)-1] == string(GenerateSecretJWT) || os.Args[len(os.Args)-1] == string(GenerateSecretJWTShort) {
		secret, err := utils.GenerateSecretJWT()
		if err != nil {
			log.Fatalf("Failed to generate secret: %v", err)
		}

		file, err := os.ReadFile(envPath)
		if err != nil {
			log.Fatalf("Failed to read .env file: %v", err)
		}

		lines := strings.Split(string(file), "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "JWT_SECRET=") {
				lines[i] = "JWT_SECRET=" + secret
				break
			}
		}

		newContent := strings.Join(lines, "\n")
		err = os.WriteFile(envPath, []byte(newContent), 0644)
		if err != nil {
			log.Fatalf("Failed to write to .env file: %v", err)
		}

		log.Println("JWT secret generated and updated in .env file successfully")

		return
	}

	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.ConnectDB()

	if os.Args[len(os.Args)-1] == string(MigrateDown) || os.Args[len(os.Args)-1] == string(MigrateDownShort) {
		err := database.MigrateDown()
		if err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migration down completed successfully")
		return
	}

	database.Migrations()

	if os.Args[len(os.Args)-1] == string(Seed) || os.Args[len(os.Args)-1] == string(SeedShort) {
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
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	err = pkg.MqttInit()
	if err != nil {
		log.Fatalf("Failed to initialize MQTT client: %v", err)
	}

	go pkg.DeviceInit(deviceRepo)

	r.Run(":8080")
}
