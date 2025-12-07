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
		apiAuth.POST("/", controllers.DeviceRegister)
		apiAuth.GET("/list", controllers.DeviceList)
		apiAuth.DELETE("/", controllers.DeviceDelete)
		apiAuth.GET("/:uuid", controllers.DeviceProfile)
		apiAuth.DELETE("/:uuid", controllers.DeviceDelete)
	}
}
