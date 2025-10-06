package services

import (
	"errors"
	"go-auth/config"
	"go-auth/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UpdateUserImage อัปเดต image_url ของ user
func (s *AuthService) UpdateUserImage(userID interface{}, imageURL string) error {
	return config.GormDB.Model(&models.User{}).Where("user_id = ?", userID).Update("image_url", imageURL).Error
}

// AuthService ให้บริการเกี่ยวกับการยืนยันตัวตน เช่น สมัครสมาชิกและเข้าสู่ระบบ
type AuthService struct {
	DBService DatabaseService
}

// NewAuthService สร้าง instance ใหม่ของ AuthService
func NewAuthService(dbService DatabaseService) *AuthService {
	return &AuthService{DBService: dbService}
}

// GetUserProfile ดึงข้อมูล user ตาม user_id
func (s *AuthService) GetUserProfile(userID interface{}) (*models.User, error) {
	var user models.User
	err := config.GormDB.First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Register ทำการสมัครสมาชิกใหม่
// รับ struct User และคืน error ถ้ามีปัญหา เช่น ข้อมูลไม่ครบหรือ user ซ้ำ
func (s *AuthService) Register(user *models.User) error {
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return ErrMissingFields
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	user.Provider = "local"
	user.Role = "user"
	user.GoogleID = nil
	if user.Firstname == "" {
		user.Firstname = ""
	}
	if user.Lastname == "" {
		user.Lastname = ""
	}
	if user.Phone == "" {
		user.Phone = ""
	}
	if user.Aboutme == "" {
		user.Aboutme = ""
	}
	if user.ImageURL == "" {
		user.ImageURL = ""
	}
	err = s.DBService.Register(user)
	if err != nil {
		if err.Error() == "user already exists" {
			return ErrUserExists
		}
		return err
	}
	return nil
}

// Login ทำการเข้าสู่ระบบและคืน user กับ token JWT
// รับ username/email และ password คืน user, token, error
func (s *AuthService) Login(usernameOrEmail, password string) (string, string, error) {
	dbUser, err := s.DBService.Login(usernameOrEmail, password)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	//create jwt token (expires in 24 hr)
	claims := jwt.MapClaims{
		"user_id": dbUser.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}
	// return role and token only
	return dbUser.Role, t, nil
}

// Custom error สำหรับ service เพื่อให้ controller ตรวจสอบและตอบกลับได้ง่าย
var (
	ErrMissingFields      = errors.New("username, email, and password are required")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
