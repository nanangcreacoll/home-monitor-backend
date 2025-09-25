package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RootRoute(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to the Home Monitor API",
		})
	})
}
