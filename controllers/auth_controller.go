package controllers

import (
	"net/http"

	"go-auth/config"
	"go-auth/models"
	"go-auth/services"

	"github.com/labstack/echo/v4"
)

var authService = services.NewAuthService()

func Register(c echo.Context) error {
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request body"})
	}

	err := authService.Register(user)
	if err == services.ErrMissingFields {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	} else if err == services.ErrUserExists {
		return c.JSON(http.StatusConflict, echo.Map{"message": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to register user"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "User registered successfully"})
}

func Login(c echo.Context) error {
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request body"})
	}

	dbUser, token, err := authService.Login(user.Username, user.Password)
	if err == services.ErrInvalidCredentials {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to login"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
		"user":  dbUser,
	})
}

// Add this function to controllers/auth_controller.go

func GetCurrentUser(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	var user models.User
	err := config.GormDB.First(&user, userID).Error
	if err != nil {
		if err.Error() == "record not found" {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Database error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"user": user})
}
