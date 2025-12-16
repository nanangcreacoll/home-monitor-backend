package routes

import (
	"home-monitor-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RootRoute(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, models.ResponseMessage{Message: "Welcome to Home Monitor API, see /api/docs for API documentation."})
	})

	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, models.ResponseMessage{Message: "Welcome to Home Monitor API, see /api/docs for API documentation."})
	})
}
