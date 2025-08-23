package main

import (
	"home-monitor-backend/config"
	"home-monitor-backend/routes"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.ConnectDB()

	r := gin.Default()

	routes.UserRoutes(r)

	r.Run(":8080")
}
