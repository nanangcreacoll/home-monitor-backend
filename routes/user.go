package routes

import (
	"home-monitor-backend/controllers"
	"home-monitor-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine, controllers *controllers.UserController) {
	api := r.Group("/api/user")
	{
		api.POST("/register", controllers.UserRegister)
		api.POST("/login", controllers.UserLogin)
	}

	apiAuth := r.Group("/api/user")
	apiAuth.Use(middlewares.Auth())
	{
		apiAuth.GET("/profile", controllers.UserProfile)
		apiAuth.PUT("/update", controllers.UserUpdate)
	}
}
