package app

import (
	"github.com/Habeebamoo/intunel-backend/internal/handlers"
	//"github.com/Habeebamoo/intunel-backend/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler,
	notificationHandler *handlers.NotificationHandler,
	jwtSecret string,
) {
	v1 := router.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	{
		auth.GET("/google", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
		auth.GET("/github", authHandler.GitHubLogin)
		auth.GET("/github/callback", authHandler.GitHubCallback)
		auth.POST("/logout", authHandler.Logout)
	}

	// Protected routes
	protected := v1.Group("/")
	//protected.Use(middlewares.AuthMiddleware(jwtSecret))
	{
		protected.POST("/notifications", notificationHandler.SendNotification)
	}
}