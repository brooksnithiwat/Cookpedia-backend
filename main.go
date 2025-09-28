package main

import (
	"go-auth/config"
	"go-auth/controllers"
	"go-auth/routes"
	"go-auth/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize GORM database
	config.InitDB()

	// Initialize SQL database for DatabaseService
	config.InitSQLDB()

	// Initialize OAuth
	config.InitOAuth()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize AuthService and AuthController
	dbService := services.NewDatabaseService(config.SQLDB)
	authService := services.NewAuthService(*dbService)
	authController := &controllers.AuthController{AuthService: authService}

	// Initialize routes with injected controller
	routes.Init(e, authController)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
