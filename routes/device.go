package routes

import (
	"home-monitor-backend/controllers"
	"home-monitor-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func DeviceRoutes(r *gin.Engine, controllers *controllers.DeviceController) {
	apiAuth := r.Group("/api/device")
	apiAuth.Use(middlewares.Auth())
	{
		apiAuth.GET("/list", controllers.DeviceList)
		apiAuth.POST("/register", controllers.DeviceRegister)
		apiAuth.DELETE("/delete", controllers.DeviceDelete)
	}
}
