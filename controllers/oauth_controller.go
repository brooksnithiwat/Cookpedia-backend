package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"go-auth/config"
	"go-auth/models"
	"net/http"
	"os"
	"time"

	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// -----------------------------
// GoogleLogin
// -----------------------------
func GoogleLogin(c echo.Context) error {
	state := generateRandomState()
	url := config.GoogleOAuthConfig.AuthCodeURL(state)
	return c.JSON(http.StatusOK, echo.Map{
		"auth_url": url,
		"state":    state,
	})
}

// -----------------------------
// GoogleCallback
// -----------------------------
func GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Authorization code not provided"})
	}

	token, err := config.GoogleOAuthConfig.Exchange(c.Request().Context(), code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Failed to exchange code for token"})
	}

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

	user, err := findOrCreateUser(googleUser)
	if err != nil {
		c.Logger().Error(err) // log ฝั่ง server
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Database error",
			"error":   err.Error(),
		})
	}

	tokenStr, err := generateJWTToken(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to generate token"})
	}

	// ✅ return JSON เสมอ
	return c.JSON(http.StatusOK, echo.Map{
		"token": tokenStr,
		"user":  user,
	})
}

func findOrCreateUser(googleUser models.GoogleUserInfo) (*models.User, error) {
	var user models.User

	// 1. หา user ด้วย google_id
	row := config.SQLDB.QueryRow(
		"SELECT user_id, username, email, google_id, provider, role FROM users WHERE google_id=$1",
		googleUser.ID,
	)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.GoogleID, &user.Provider, &user.Role)
	if err == nil {
		_, err = config.SQLDB.Exec(
			"UPDATE users SET username=$1, email=$2 WHERE user_id=$3",
			googleUser.Name, googleUser.Email, user.ID,
		)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// 2. หา user ด้วย email
	row = config.SQLDB.QueryRow(
		"SELECT user_id, username, email, google_id, provider, role FROM users WHERE email=$1",
		googleUser.Email,
	)
	err = row.Scan(&user.ID, &user.Username, &user.Email, &user.GoogleID, &user.Provider, &user.Role)
	if err == nil {
		_, err = config.SQLDB.Exec(
			"UPDATE users SET google_id=$1 WHERE user_id=$2",
			googleUser.ID, user.ID,
		)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// 3. สร้าง user ใหม่
	var newID int
	err = config.SQLDB.QueryRow(
		"INSERT INTO users(username,email,google_id,provider,role) VALUES($1,$2,$3,$4,$5) RETURNING user_id",
		googleUser.Name, googleUser.Email, googleUser.ID, "google", "user",
	).Scan(&newID)
	if err != nil {
		return nil, err
	}

	user = models.User{
		ID:       newID,
		Username: googleUser.Name,
		Email:    googleUser.Email,
		GoogleID: &googleUser.ID,
		Provider: "google",
		Role:     "user",
	}

	return &user, nil
}

// -----------------------------
// JWT
// -----------------------------
func generateJWTToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}

// -----------------------------
// Random state for OAuth
// -----------------------------
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
