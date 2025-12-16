package middlewares

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	host := os.Getenv("APP_HOST")
	allowAll := false

	if host == "" {
		host = "http://localhost"
		allowAll = true
	} else if strings.HasPrefix(host, "http://localhost") && strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		allowAll = true
	}

	if !allowAll {
		config := cors.Config{
			AllowOrigins:     []string{host},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}

		return cors.New(config)
	}

	config := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}

	return cors.New(config)
}
