package controllers

import (
	"fmt"
	"go-auth/models"
	"go-auth/supabaseutil"
	"net/http"
	"strconv"
	"strings"
	"github.com/labstack/echo/v4"
)

func (ac *AuthController) UpdateUserProfile(c echo.Context) error {
	// 1) ตรวจสอบ user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": "User not authenticated",
		})
	}

	// แปลง user_id
	uidStr := fmt.Sprintf("%v", uid)
	userID, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user_id"})
	}

	// ดึง form params ทั้งหมด
	formParams, err := c.FormParams()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid form data"})
	}

	// รับค่าจาก form (แต่ต้องเช็คว่า key ถูกส่งมามั้ย)
	fields := []string{}
	values := []interface{}{}

	if _, ok := formParams["firstname"]; ok {
		fields = append(fields, "firstname")
		values = append(values, c.FormValue("firstname")) // จะได้ทั้ง "" หรือค่าอื่น ๆ
	}
	if _, ok := formParams["lastname"]; ok {
		fields = append(fields, "lastname")
		values = append(values, c.FormValue("lastname"))
	}
	if _, ok := formParams["email"]; ok {
		fields = append(fields, "email")
		values = append(values, c.FormValue("email"))
	}
	if _, ok := formParams["phone"]; ok {
		fields = append(fields, "phone")
		values = append(values, c.FormValue("phone"))
	}
	if _, ok := formParams["aboutme"]; ok {
		fields = append(fields, "aboutme")
		values = append(values, c.FormValue("aboutme"))
	}

	// จัดการไฟล์รูป
	file, err := c.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to read image",
			"error":   err.Error(),
		})
	}

	if file != nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to open image"})
		}
		defer src.Close()

		imageURL, err := supabaseutil.UploadFile(src, file, userID, "ProfileImage/profile")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Failed to upload image",
				"error":   err.Error(),
			})
		}

		fields = append(fields, "image_url")
		values = append(values, imageURL)
	}

	// ถ้าไม่มีอะไรส่งมาเลย
	if len(fields) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "No data to update"})
	}

	// where condition
	whereCon := fmt.Sprintf("user_id = $%d", len(fields)+1)
	values = append(values, userID)

	// update
	rowsAffected, err := ac.AuthService.DBService.UpdateData("users", fields, whereCon, values)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to update user profile", "error": err.Error(),
		})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Edit Profile Success"})
}

// splitAndTrim แยก string ด้วย sep แล้ว trim space
func splitAndTrim(s, sep string) []string {
	arr := []string{}
	for _, v := range strings.Split(s, sep) {
		v = strings.TrimSpace(v)
		if v != "" {
			arr = append(arr, v)
		}
	}
	return arr
}
func (ac *AuthController) GetUserProfileByID(c echo.Context) error {
	// ดึง user id จาก path เช่น /userprofile/5
	userID := c.Param("id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "User ID is required"})
	}

	// ฟิลด์ที่อยาก select
	fields := []string{"user_id", "firstname", "lastname", "email", "phone", "aboutme", "image_url"}
	whereCon := "user_id = ?"
	whereArgs := []interface{}{userID}

	// ดึงข้อมูลจาก DB
	results, err := ac.AuthService.DBService.SelectData("users", fields, true, whereCon, whereArgs, false, "", "", "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Database error",
			"error":   err.Error(),
		})
	}
	if len(results) == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	}

	row := results[0]
	profile := models.UserProfile{
		UserID:    row["user_id"].(int64),
		Firstname: row["firstname"].(string),
		Lastname:  row["lastname"].(string),
		Email:     row["email"].(string),
		Phone:     row["phone"].(string),
		AboutMe:   row["aboutme"].(string),
		ImageURL:  row["image_url"].(string),
	}

	return c.JSON(http.StatusOK, profile)
}
func (ac *AuthController) GetAllUser(c echo.Context) error {
	// เลือก field ที่อยากได้
	fields := []string{"user_id", "username", "firstname", "lastname", "email", "phone", "aboutme", "image_url"}

	// ดึงข้อมูลทั้งหมด ไม่มี where condition
	results, err := ac.AuthService.DBService.SelectData("users", fields, true, "", nil, false, "", "", "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Database error",
			"error":   err.Error(),
		})
	}

	// แปลงเป็น slice ของ UserProfile
	users := make([]models.UserProfile, 0, len(results))
	for _, row := range results {
		uid, ok := row["user_id"].(int64) // หรือ int ตาม DB
		if !ok {
			uid = 0
		}

		user := models.UserProfile{
			UserID:    uid,
			UserName:  row["username"].(string),
			Firstname: row["firstname"].(string),
			Lastname:  row["lastname"].(string),
			Email:     row["email"].(string),
			Phone:     row["phone"].(string),
			AboutMe:   row["aboutme"].(string),
			ImageURL:  row["image_url"].(string),
		}
		users = append(users, user)
	}

	return c.JSON(http.StatusOK, users)
}

func (ac *AuthController) GetUserProfile(c echo.Context) error {
	// ดึง user_id จาก context (เช่น JWT)
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	fields := []string{"user_id", "username", "firstname", "lastname", "email", "phone", "aboutme", "image_url"}
	whereCon := "user_id = ?"
	whereArgs := []interface{}{userID}

	results, err := ac.AuthService.DBService.SelectData("users", fields, true, whereCon, whereArgs, false, "", "", "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Database error", "error": err.Error()})
	}
	if len(results) == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	}

	row := results[0]

	profile := models.UserProfile{
		UserID:    row["user_id"].(int64), // ใช้ int64 ตรงๆ
		UserName:  row["username"].(string),
		Firstname: row["firstname"].(string),
		Lastname:  row["lastname"].(string),
		Email:     row["email"].(string),
		Phone:     row["phone"].(string),
		AboutMe:   row["aboutme"].(string),
		ImageURL:  row["image_url"].(string),
	}

	return c.JSON(http.StatusOK, profile)
}
