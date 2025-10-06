package controllers

import (
	"net/http"

	"go-auth/models"
	"go-auth/services"

	"github.com/labstack/echo/v4"
)

// POST /api/userprofile (multipart/form-data: image)

type AuthController struct {
	AuthService *services.AuthService
}

func (ac *AuthController) Register(c echo.Context) error {
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request body"})
	}

	err := ac.AuthService.Register(user)
	if err == services.ErrMissingFields {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	} else if err == services.ErrUserExists {
		return c.JSON(http.StatusConflict, echo.Map{"message": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to register user"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "User registered successfully"})
}

func (ac *AuthController) Login(c echo.Context) error {
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request body"})
	}

	dbUser, token, err := ac.AuthService.Login(user.Username, user.Password)
	if err == services.ErrInvalidCredentials {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to login"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
		"Role":  dbUser,
	})
}
