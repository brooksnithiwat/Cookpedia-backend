package controllers

import (
	"crypto/rand"
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
	"gorm.io/gorm"
)

// GoogleLogin: เริ่ม OAuth flow และคืน URL ให้ frontend
func GoogleLogin(c echo.Context) error {
	state := generateRandomState()
	url := config.GoogleOAuthConfig.AuthCodeURL(state)
	return c.JSON(http.StatusOK, echo.Map{
		"auth_url": url,
		"state":    state,
	})
}

// GoogleCallback: รับ callback จาก Google, ดึง user info, login/signup, ออก token
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
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Database error",
			"error":   err.Error(), // ✅ ส่งรายละเอียด error กลับมาด้วย
		})
	}

	tokenStr, err := generateJWTToken(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to generate token"})
	}

	// ✅ ไม่ redirect แล้ว → return JSON ตลอด
	return c.JSON(http.StatusOK, echo.Map{
		"token": tokenStr,
		"user":  user,
	})
}

// findOrCreateUser: หา user จาก google_id/email หรือสร้างใหม่
func findOrCreateUser(googleUser models.GoogleUserInfo) (*models.User, error) {
	var user models.User

	// 1. หา user ด้วย google_id
	err := config.GormDB.Where("google_id = ?", googleUser.ID).First(&user).Error
	if err == nil {
		// update ข้อมูลล่าสุดจาก Google
		user.Email = googleUser.Email
		user.Username = googleUser.Name
		if saveErr := config.GormDB.Save(&user).Error; saveErr != nil {
			return nil, saveErr
		}
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. หา user ด้วย email
	err = config.GormDB.Where("email = ?", googleUser.Email).First(&user).Error
	if err == nil {
		user.GoogleID = &googleUser.ID
		if saveErr := config.GormDB.Save(&user).Error; saveErr != nil {
			return nil, saveErr
		}
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 3. ถ้าไม่เจอ → สร้างใหม่
	user = models.User{
		Username: googleUser.Name,
		Email:    googleUser.Email,
		GoogleID: &googleUser.ID,
		Provider: "google",
		Role:     "user",
	}
	if err := config.GormDB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// generateJWTToken: สร้าง JWT token สำหรับ user
func generateJWTToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}

// generateRandomState: สุ่ม state สำหรับ OAuth
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
