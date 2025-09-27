// Package services รวม business logic ที่แยกออกจาก controller
package services

import (
	"errors"
	"os"
	"time"

	"go-auth/config"
	"go-auth/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService ให้บริการเกี่ยวกับการยืนยันตัวตน เช่น สมัครสมาชิกและเข้าสู่ระบบ
type AuthService struct{}

// NewAuthService สร้าง instance ใหม่ของ AuthService
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Register ทำการสมัครสมาชิกใหม่
// รับ struct User และคืน error ถ้ามีปัญหา เช่น ข้อมูลไม่ครบหรือ user ซ้ำ
func (s *AuthService) Register(user *models.User) error {
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return ErrMissingFields
	}

	var existing models.User
	err := config.GormDB.Where("username = ? OR email = ?", user.Username, user.Email).First(&existing).Error
	if err == nil {
		return ErrUserExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	user.Password = string(hashedPassword)
	user.Provider = "local"
	user.GoogleID = nil // set NULL ใน DB
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := config.GormDB.Create(user).Error; err != nil {
		return err
	}

	return nil
}

// Login ทำการเข้าสู่ระบบและคืน user กับ token JWT
// รับ username/email และ password คืน user, token, error
func (s *AuthService) Login(usernameOrEmail, password string) (*models.User, string, error) {
	var dbUser models.User
	err := config.GormDB.Where("(username = ? OR email = ?) AND provider = ?", usernameOrEmail, usernameOrEmail, "local").First(&dbUser).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"user_id": dbUser.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, "", err
	}

	dbUser.Password = "" // ไม่คืน password
	return &dbUser, t, nil
}

// Custom error สำหรับ service เพื่อให้ controller ตรวจสอบและตอบกลับได้ง่าย
var (
	ErrMissingFields      = errors.New("username, email, and password are required")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
