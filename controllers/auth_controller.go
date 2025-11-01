package controllers

import (
	"net/http"
  	"strings"
	"go-auth/models"
	"go-auth/services"
	"net/http"
	"strings"

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
	if len(user.Username) == 0 || len(user.Password) == 0 || len(user.Email) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Username, password, and email are required"})
	}
	if len(user.Username) > 30 {
		return c.JSON(http.StatusBadRequest, echo.Map{"Username": "Username must not exceed 30 characters"})
	}
	if len(user.Password) > 30 {
		return c.JSON(http.StatusBadRequest, echo.Map{"Password": "Password must not exceed 30 characters"})
	}
	if len(user.Email) > 30 {
		return c.JSON(http.StatusBadRequest, echo.Map{"Email": "Email must not exceed 30 characters"})
	}
	if !strings.Contains(user.Email, "@") {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid email format: must contain '@'",
		})
	}

	err := ac.AuthService.Register(user)
	if err == services.ErrMissingFields {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	} else if err == services.ErrUserExists {
		return c.JSON(http.StatusConflict, echo.Map{"message": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to register user",
			"error":   err.Error()})
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
