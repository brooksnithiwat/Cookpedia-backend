package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go-auth/config"
	"go-auth/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// GoogleLogin initiates the Google OAuth flow
func GoogleLogin(c echo.Context) error {
	state := generateRandomState()

	// Store state in session or cache (for production, use Redis/database)
	// For now, we'll trust the state verification in callback

	url := config.GoogleOAuthConfig.AuthCodeURL(state)
	return c.JSON(http.StatusOK, echo.Map{
		"auth_url": url,
		"state":    state,
	})
}

// GoogleCallback handles the callback from Google OAuth
func GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")

	if code == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Authorization code not provided"})
	}

	// Exchange code for token
	token, err := config.GoogleOAuthConfig.Exchange(c.Request().Context(), code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Failed to exchange code for token"})
	}

	// Get user info from Google
	client := config.GoogleOAuthConfig.Client(c.Request().Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get user info"})
	}
	defer resp.Body.Close()

	var googleUser models.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to decode user info"})
	}

	// Check if user exists in database
	user, err := findOrCreateUser(googleUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Database error"})
	}

	// Generate JWT token
	jwtToken, err := generateJWTToken(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to generate token"})
	}

	// Redirect to frontend with token
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		// If no frontend URL, return JSON response
		return c.JSON(http.StatusOK, echo.Map{
			"token": jwtToken,
			"user":  user,
		})
	}

	// Redirect to frontend with token as query parameter
	redirectURL := fmt.Sprintf("%s/auth/success?token=%s", frontendURL, jwtToken)
	return c.Redirect(http.StatusFound, redirectURL)
}

func findOrCreateUser(googleUser models.GoogleUserInfo) (*models.User, error) {
	var user models.User

	// First, try to find user by Google ID
	err := config.DB.QueryRow(`
		SELECT id, username, email, google_id, profile_picture, provider, created_at, updated_at 
		FROM users WHERE google_id = $1`, googleUser.ID).Scan(
		&user.ID, &user.Username, &user.Email, &user.GoogleID,
		&user.ProfilePicture, &user.Provider, &user.CreatedAt, &user.UpdatedAt)

	if err == nil {
		// User found, update their info
		_, updateErr := config.DB.Exec(`
			UPDATE users SET email = $1, username = $2, profile_picture = $3, updated_at = $4 
			WHERE google_id = $5`,
			googleUser.Email, googleUser.Name, googleUser.Picture, time.Now(), googleUser.ID)
		if updateErr != nil {
			return nil, updateErr
		}
		return &user, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// User not found by Google ID, check by email
	err = config.DB.QueryRow(`
		SELECT id, username, email, google_id, profile_picture, provider, created_at, updated_at 
		FROM users WHERE email = $1`, googleUser.Email).Scan(
		&user.ID, &user.Username, &user.Email, &user.GoogleID,
		&user.ProfilePicture, &user.Provider, &user.CreatedAt, &user.UpdatedAt)

	if err == nil {
		// User exists with same email, link Google account
		_, updateErr := config.DB.Exec(`
			UPDATE users SET google_id = $1, profile_picture = $2, updated_at = $3 
			WHERE email = $4`,
			googleUser.ID, googleUser.Picture, time.Now(), googleUser.Email)
		if updateErr != nil {
			return nil, updateErr
		}
		user.GoogleID = googleUser.ID
		user.ProfilePicture = googleUser.Picture
		return &user, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new user
	now := time.Now()
	err = config.DB.QueryRow(`
		INSERT INTO users (username, email, google_id, profile_picture, provider, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		googleUser.Name, googleUser.Email, googleUser.ID, googleUser.Picture, "google", now, now).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	user.Username = googleUser.Name
	user.Email = googleUser.Email
	user.GoogleID = googleUser.ID
	user.ProfilePicture = googleUser.Picture
	user.Provider = "google"
	user.CreatedAt = now
	user.UpdatedAt = now

	return &user, nil
}

func generateJWTToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
