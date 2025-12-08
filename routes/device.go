package routes

import (
	"home-monitor-backend/controllers"
	"home-monitor-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func DeviceRoutes(r *gin.Engine, controllers *controllers.DeviceController) {
	apiAuth := r.Group("/api/device")

	apiAuth.GET("/mac/:mac_address", controllers.DeviceProfileByMacAddress)

	apiAuth.Use(middlewares.Auth())
	{
		apiAuth.POST("/", controllers.DeviceRegister)
		apiAuth.GET("/list", controllers.DeviceList)
		apiAuth.DELETE("/", controllers.DeviceDelete)
		apiAuth.GET("/:uuid", controllers.DeviceProfile)
		apiAuth.PUT("/:uuid", controllers.DeviceUpdate)
		apiAuth.DELETE("/:uuid", controllers.DeviceDelete)
		apiAuth.GET("/measurements", controllers.DeviceMeasurements)
		apiAuth.DELETE("/measurements", controllers.DeviceDeleteMeasurements)
		apiAuth.POST("/:uuid/measurements", controllers.DeviceCreateMeasurement)
	}
}
