package routes

import (
	"home-monitor-backend/controllers"
	"home-monitor-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine, controllers *controllers.UserController) {
	api := r.Group("/api/user")
	{
		api.POST("/login", controllers.UserLogin)
	}

	apiAuth := r.Group("/api/user")
	apiAuth.Use(middlewares.Auth())
	{
		apiAuth.POST("", controllers.UserRegister)
		apiAuth.GET("", controllers.UserProfile)
		apiAuth.PUT("", controllers.UserUpdate)
		apiAuth.DELETE("", controllers.UserDelete)
		apiAuth.GET("/list", controllers.UserList)
	}
}
