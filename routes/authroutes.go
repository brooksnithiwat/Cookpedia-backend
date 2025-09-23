package routes

import (
	"go-auth/controllers"
	jwtmiddleware "go-auth/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Init(e *echo.Echo) {
	// CORS middleware for frontend integration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // In production, specify your frontend domain
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Traditional auth routes
	e.POST("/register", controllers.Register)
	e.POST("/login", controllers.Login)

	// Google OAuth routes
	e.GET("/auth/google", controllers.GoogleLogin)
	e.GET("/auth/google/callback", controllers.GoogleCallback)

	// Protected routes
	api := e.Group("/api")
	api.Use(jwtmiddleware.JWTMiddleware())

	api.GET("/profile", func(c echo.Context) error {
		userID := c.Get("user_id")
		return c.JSON(200, echo.Map{
			"message": "You are logged in",
			"user_id": userID,
		})
	})

	// Get current user info
	api.GET("/me", controllers.GetCurrentUser)
}
