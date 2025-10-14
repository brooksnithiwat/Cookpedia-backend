package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"

	"fmt"
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

	// แลก code เป็น access token ของ Google
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

	// หา หรือ สร้าง user ใน DB
	dbUser, err := findOrCreateUser(googleUser)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Database error",
			"error":   err.Error(),
		})
	}

	// generate JWT token แบบเดียวกับ Login ปกติ
	tokenStr, err := generateJWTToken(dbUser) // note: ต้องแก้ generateJWTToken ให้รับ *models.User
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to generate token"})
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL != "" {
		// redirect ไป frontend พร้อม token และ role
		redirectURL := fmt.Sprintf("%s/auth/success?token=%s&role=%s", frontendURL, tokenStr, dbUser.Role)
		return c.Redirect(http.StatusFound, redirectURL)
	}

	// ถ้าไม่มี frontend URL → return JSON แบบ unified
	return c.JSON(http.StatusOK, echo.Map{
		"token": tokenStr,
		"user": map[string]interface{}{
			"id":       dbUser.ID,
			"username": dbUser.Username,
			"email":    dbUser.Email,
			"role":     dbUser.Role,
		},
	})
}

func findOrCreateUser(googleUser models.GoogleUserInfo) (*models.User, error) {
	var user models.User

	// 1. หา user ด้วย google_id
	row := config.SQLDB.QueryRow(
		"SELECT user_id, username, email, google_id, provider, role, firstname, lastname, phone, aboutme, image_url FROM users WHERE google_id=$1",
		googleUser.ID,
	)
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.GoogleID, &user.Provider,
		&user.Role, &user.Firstname, &user.Lastname, &user.Phone, &user.Aboutme, &user.ImageURL,
	)
	if err == nil {
		// อัปเดตเฉพาะ username, email (ไม่แตะ image_url)
		_, err = config.SQLDB.Exec(
			"UPDATE users SET username=$1, email=$2 WHERE user_id=$3",
			googleUser.Name, googleUser.Email, user.ID,
		)
		if err != nil {
			return nil, err
		}

		user.Username = googleUser.Name
		user.Email = googleUser.Email

		// ถ้ายังไม่มี image_url ให้เติมรอบแรกเท่านั้น
		if user.ImageURL == "" {
			_, err = config.SQLDB.Exec(
				"UPDATE users SET image_url=$1 WHERE user_id=$2",
				googleUser.ImageURL, user.ID,
			)
			if err == nil {
				user.ImageURL = googleUser.ImageURL
			}
		}

		return &user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// 2. หา user ด้วย email
	row = config.SQLDB.QueryRow(
		"SELECT user_id, username, email, google_id, provider, role, firstname, lastname, phone, aboutme, image_url FROM users WHERE email=$1",
		googleUser.Email,
	)
	err = row.Scan(
		&user.ID, &user.Username, &user.Email, &user.GoogleID, &user.Provider,
		&user.Role, &user.Firstname, &user.Lastname, &user.Phone, &user.Aboutme, &user.ImageURL,
	)
	if err == nil {
		// อัปเดตเฉพาะ google_id (ไม่แตะ image_url)
		_, err = config.SQLDB.Exec(
			"UPDATE users SET google_id=$1 WHERE user_id=$2",
			googleUser.ID, user.ID,
		)
		if err != nil {
			return nil, err
		}

		user.GoogleID = &googleUser.ID

		// ถ้ายังไม่มี image_url ให้เติมรอบแรกเท่านั้น
		if user.ImageURL == "" {
			_, err = config.SQLDB.Exec(
				"UPDATE users SET image_url=$1 WHERE user_id=$2",
				googleUser.ImageURL, user.ID,
			)
			if err == nil {
				user.ImageURL = googleUser.ImageURL
			}
		}

		return &user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// 3. สร้าง user ใหม่ พร้อม default field แบบ Register
	var newID int
	err = config.SQLDB.QueryRow(
		`INSERT INTO users(
			username,email,password,firstname,lastname,phone,aboutme,image_url,google_id,provider,role
		) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING user_id`,
		googleUser.Name,     // username
		googleUser.Email,    // email
		"",                  // password
		"",                  // firstname
		"",                  // lastname
		"",                  // phone
		"",                  // aboutme
		googleUser.ImageURL, // image_url
		googleUser.ID,       // google_id
		"google",            // provider
		"user",              // role
	).Scan(&newID)
	if err != nil {
		return nil, err
	}

	user = models.User{
		ID:        newID,
		Username:  googleUser.Name,
		Email:     googleUser.Email,
		Password:  "",
		Firstname: "",
		Lastname:  "",
		Phone:     "",
		Aboutme:   "",
		ImageURL:  googleUser.ImageURL,
		GoogleID:  &googleUser.ID,
		Provider:  "google",
		Role:      "user",
	}

	return &user, nil
}

// -----------------------------
// JWT
// -----------------------------
func generateJWTToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
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
