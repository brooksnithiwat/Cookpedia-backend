package controllers

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"go-auth/config"
	"go-auth/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func Register(c echo.Context) error {
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request body"})
	}

	// Validate required fields
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Username, email, and password are required"})
	}

	// Check if user already exists
	var existingID int
	err := config.DB.QueryRow("SELECT id FROM users WHERE username = $1 OR email = $2", user.Username, user.Email).Scan(&existingID)
	if err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"message": "User already exists"})
	} else if err != sql.ErrNoRows {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Database error"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to hash password"})
	}

	now := time.Now()
	_, err = config.DB.Exec(`
		INSERT INTO users (username, email, password, provider, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		user.Username, user.Email, string(hashedPassword), "local", now, now)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to register user"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "User registered successfully"})
}

func Login(c echo.Context) error {
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request body"})
	}

	var dbUser models.User
	err := config.DB.QueryRow(`
		SELECT id, username, email, password, provider 
		FROM users WHERE (username = $1 OR email = $1) AND provider = 'local'`,
		user.Username).Scan(&dbUser.ID, &dbUser.Username, &dbUser.Email, &dbUser.Password, &dbUser.Provider)

	if err == sql.ErrNoRows {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid credentials"})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Database error"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid credentials"})
	}

	claims := jwt.MapClaims{
		"user_id": dbUser.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to generate token"})
	}

	// Return user info along with token (exclude password)
	dbUser.Password = ""
	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
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
	err := config.DB.QueryRow(`
		SELECT id, username, email, google_id, profile_picture, provider, created_at, updated_at 
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.GoogleID,
		&user.ProfilePicture, &user.Provider, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Database error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"user": user})
}
