package routes

import (
	"home-monitor-backend/controllers"
	"home-monitor-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func DeviceRoutes(r *gin.Engine, controllers *controllers.DeviceController) {
	api := r.Group("/api/device")
	{
		api.POST("/login", controllers.DeviceLogin)
	}

	apiAuth := r.Group("/api/device")
	apiAuth.Use(middlewares.Auth())
	{
		apiAuth.POST("", controllers.DeviceRegister)
		apiAuth.GET("/list", controllers.DeviceList)
		apiAuth.DELETE("", controllers.DeviceDelete)
		apiAuth.GET("/:uuid", controllers.DeviceProfile)
		apiAuth.PUT("/:uuid", controllers.DeviceUpdate)
		apiAuth.PATCH("/:uuid/status", controllers.DeviceUpdateStatus)
		apiAuth.DELETE("/:uuid", controllers.DeviceDelete)
		apiAuth.GET("/measurements", controllers.DeviceMeasurements)
		apiAuth.DELETE("/measurements", controllers.DeviceDeleteMeasurements)
		apiAuth.POST("/:uuid/measurements", controllers.DeviceCreateMeasurement)
	}
}
