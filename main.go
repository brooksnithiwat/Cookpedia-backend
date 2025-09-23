package main

import (
	"go-auth/config"
	"go-auth/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize database
	config.InitDB()
	
	// Initialize OAuth
	config.InitOAuth()
	
	// Create Echo instance
	e := echo.New()
	
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// Initialize routes
	routes.Init(e)
	
	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}