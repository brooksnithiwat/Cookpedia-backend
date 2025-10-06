package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	// "fmt"
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
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Database error",
			"error":   err.Error(),
		})
	}

	tokenStr, err := generateJWTToken(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to generate token"})
	}

	// ✅ return JSON เหมือน login ปกติ
	return c.JSON(http.StatusOK, echo.Map{
		"token": tokenStr,
		"user":  user,
	})
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

// func findOrCreateUser(googleUser models.GoogleUserInfo) (*models.User, error) {
// 	var user models.User
// 	var googleID, imageURL sql.NullString

// 	// 1. หา user ด้วย google_id
// 	row := config.SQLDB.QueryRow(
// 		"SELECT user_id, username, email, google_id, image_url, provider, role FROM users WHERE google_id=$1",
// 		googleUser.ID,
// 	)
// 	err := row.Scan(&user.ID, &user.Username, &user.Email, &googleID, &imageURL, &user.Provider, &user.Role)
// 	if err == nil {
// 		if googleID.Valid {
// 			user.GoogleID = &googleID.String
// 		}
// 		if imageURL.Valid {
// 			user.ImageURL = imageURL.String
// 		}
// 		// อัปเดตชื่อ, อีเมล และรูปจาก Google
// 		_, err = config.SQLDB.Exec(
// 			"UPDATE users SET username=$1, email=$2, google_id=$3, image_url=$4 WHERE user_id=$5",
// 			googleUser.Name, googleUser.Email, googleUser.ID, googleUser.ImageURL, user.ID,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		user.GoogleID = &googleUser.ID
// 		user.ImageURL = googleUser.ImageURL
// 		return &user, nil
// 	}
// 	if !errors.Is(err, sql.ErrNoRows) {
// 		return nil, err
// 	}

// 	// 2. หา user ด้วย email
// 	row = config.SQLDB.QueryRow(
// 		"SELECT user_id, username, email, google_id, image_url, provider, role FROM users WHERE email=$1",
// 		googleUser.Email,
// 	)
// 	err = row.Scan(&user.ID, &user.Username, &user.Email, &googleID, &imageURL, &user.Provider, &user.Role)
// 	if err == nil {
// 		if googleID.Valid {
// 			user.GoogleID = &googleID.String
// 		}
// 		if imageURL.Valid {
// 			user.ImageURL = imageURL.String
// 		}
// 		// bind google_id + image_url ให้ user เดิม
// 		_, err = config.SQLDB.Exec(
// 			"UPDATE users SET google_id=$1, image_url=$2, provider=$3 WHERE user_id=$4",
// 			googleUser.ID, googleUser.ImageURL, "google", user.ID,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		user.GoogleID = &googleUser.ID
// 		user.ImageURL = googleUser.ImageURL
// 		user.Provider = "google"
// 		return &user, nil
// 	}
// 	if !errors.Is(err, sql.ErrNoRows) {
// 		return nil, err
// 	}

// 	// 3. ถ้าไม่เจอ → สร้าง user ใหม่
// 	var newID int
// 	err = config.SQLDB.QueryRow(
// 		"INSERT INTO users(username,email,google_id,image_url,provider,role) VALUES($1,$2,$3,$4,$5,$6) RETURNING user_id",
// 		googleUser.Name, googleUser.Email, googleUser.ID, googleUser.ImageURL, "google", "user",
// 	).Scan(&newID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	user = models.User{
// 		ID:       newID,
// 		Username: googleUser.Name,
// 		Email:    googleUser.Email,
// 		GoogleID: &googleUser.ID,
// 		ImageURL: googleUser.ImageURL,
// 		Provider: "google",
// 		Role:     "user",
// 	}

// 	return &user, nil
// }

func findOrCreateUser(googleUser models.GoogleUserInfo) (*models.User, error) {
	var user models.User
	var googleID sql.NullString

	// 1. หา user ด้วย google_id
	row := config.SQLDB.QueryRow(
		"SELECT user_id, username, email, google_id, image_url, provider, role FROM users WHERE google_id=$1",
		googleUser.ID,
	)
	var imageURL string
	err := row.Scan(&user.ID, &user.Username, &user.Email, &googleID, &imageURL, &user.Provider, &user.Role)
	if err == nil {
		if googleID.Valid {
			user.GoogleID = &googleID.String
		}
		user.ImageURL = imageURL // ไม่อัปเดตรูป
		// update ชื่อ + email เท่านั้น
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
		"SELECT user_id, username, email, google_id, image_url, provider, role FROM users WHERE email=$1",
		googleUser.Email,
	)
	err = row.Scan(&user.ID, &user.Username, &user.Email, &googleID, &imageURL, &user.Provider, &user.Role)
	if err == nil {
		if googleID.Valid {
			user.GoogleID = &googleID.String
		}
		user.ImageURL = imageURL // ไม่อัปเดตรูป
		// bind google_id ให้ user เดิม แต่ไม่อัปเดตรูป
		_, err = config.SQLDB.Exec(
			"UPDATE users SET google_id=$1, provider=$2 WHERE user_id=$3",
			googleUser.ID, "google", user.ID,
		)
		if err != nil {
			return nil, err
		}
		user.GoogleID = &googleUser.ID
		user.Provider = "google"
		return &user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// 3. ถ้าไม่เจอ → สร้าง user ใหม่ พร้อม image_url
	var newID int
	err = config.SQLDB.QueryRow(
		"INSERT INTO users(username,email,google_id,image_url,provider,role) VALUES($1,$2,$3,$4,$5,$6) RETURNING user_id",
		googleUser.Name, googleUser.Email, googleUser.ID, googleUser.ImageURL, "google", "user",
	).Scan(&newID)
	if err != nil {
		return nil, err
	}

	user = models.User{
		ID:       newID,
		Username: googleUser.Name,
		Email:    googleUser.Email,
		GoogleID: &googleUser.ID,
		ImageURL: googleUser.ImageURL,
		Provider: "google",
		Role:     "user",
	}

	return &user, nil
}
