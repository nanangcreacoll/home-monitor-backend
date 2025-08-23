package routes

import (
	"home-monitor-backend/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/user/register", controllers.UserRegister)
		api.POST("/user/login", controllers.UserLogin)
		api.GET("/user/profile", controllers.UserProfile)
	}
}
