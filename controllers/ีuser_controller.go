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

func (ac *AuthController) CreatePost(c echo.Context) error {
	// 1) ตรวจสอบ user_id จาก token
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": "User not authenticated",
		})
	}

	// 2) อ่านค่า form
	menuName := c.FormValue("menu_name")
	story := c.FormValue("story")
	categoriesStr := c.FormValue("categories")     // ตัวอย่าง: "One-dish|Dessert"
	ingredientsStr := c.FormValue("ingredients")   // ตัวอย่าง: "egg|rice|pork"
	instructionsStr := c.FormValue("instructions") // ตัวอย่าง: "step1|step2|step3"

	if menuName == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "menu_name is required"})
	}

	// split string เป็น slice
	var categories, ingredients, instructions []string
	if categoriesStr != "" {
		categories = splitAndTrim(categoriesStr, "|")
	}
	if ingredientsStr != "" {
		ingredients = splitAndTrim(ingredientsStr, "|")
	}
	if instructionsStr != "" {
		instructions = splitAndTrim(instructionsStr, "|")
	}

	// 3) อัปโหลดไฟล์ถ้ามี
	var imageURL string
	file, err := c.FormFile("image")
	if file != nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to open image"})
		}
		defer src.Close()

		imageURL, err = supabaseutil.UploadFile(src, file, userID, "PostImage/post")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Failed to upload image", "error": err.Error(),
			})
		}
	}

	// 4) lookup category_tag_id จากชื่อ (ใช้ตัวแรกเป็นหลัก)
	var categoryTagID int
	if len(categories) > 0 {
		err := ac.AuthService.DBService.DB.QueryRow("SELECT category_tag_id FROM categories_tag WHERE category_tag_name = $1", categories[0]).Scan(&categoryTagID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid category name: " + categories[0]})
		}
	}

	// 5) insert post ลง DB ด้วย InsertData
	postData := map[string]interface{}{
		"user_id":   userID,
		"menu_name": menuName,
		"story":     story,
		"image_url": imageURL,
	}
	_, err = ac.AuthService.DBService.InsertData("posts", postData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to create post", "error": err.Error(),
		})
	}

	// ดึง post_id ล่าสุด (อาจต้องแก้ให้เหมาะกับ production)
	var postID int
	err = ac.AuthService.DBService.DB.QueryRow("SELECT currval(pg_get_serial_sequence('posts','post_id'))").Scan(&postID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get post_id", "error": err.Error()})
	}

	// 6) insert post_categories (รองรับหลาย category)
	for _, cat := range categories {
		var catID int
		err := ac.AuthService.DBService.DB.QueryRow("SELECT category_tag_id FROM categories_tag WHERE category_tag_name = $1", cat).Scan(&catID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid category name: " + cat})
		}
		_, err = ac.AuthService.DBService.InsertData("post_categories", map[string]interface{}{
			"post_id":         postID,
			"category_tag_id": catID,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert post_categories", "error": err.Error()})
		}
	}

	// 6) insert ingredients
	for _, ing := range ingredients {
		ingData := map[string]interface{}{
			"post_id": postID,
			"detail":  ing,
		}
		_, err := ac.AuthService.DBService.InsertData("ingredients_detail", ingData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Failed to insert ingredient", "error": err.Error(),
			})
		}
	}

	// 7) insert instructions (มี step_number)
	for i, step := range instructions {
		insData := map[string]interface{}{
			"post_id":     postID,
			"step_number": i + 1,
			"detail":      step,
		}
		_, err := ac.AuthService.DBService.InsertData("instructions", insData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Failed to insert instruction", "error": err.Error(),
			})
		}
	}

	// 8) return response
	return c.JSON(http.StatusCreated, echo.Map{
		"message":      "Post created successfully",
		"post_id":      postID,
		"menu_name":    menuName,
		"story":        story,
		"image_url":    imageURL,
		"ingredients":  ingredients,
		"instructions": instructions,
	})
}

func (ac *AuthController) UpdateUserProfile(c echo.Context) error {
	// 1) ตรวจสอบ user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": "User not authenticated",
		})
	}

	// แปลง user_id ให้เป็น string ก่อนแล้วค่อย parse เป็น int64
	uidStr := fmt.Sprintf("%v", uid)
	userID, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user_id"})
	}

	// รับค่า form-data อื่น ๆ
	firstname := c.FormValue("firstname")
	lastname := c.FormValue("lastname")
	phone := c.FormValue("phone")
	email := c.FormValue("email")
	aboutme := c.FormValue("aboutme")

	file, err := c.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to read image",
			"error":   err.Error(),
		})
	}

	var imageURL string
	if file != nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to open image"})
		}
		defer src.Close()

		imageURL, err = supabaseutil.UploadFile(src, file, userID, "ProfileImage/profile")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "Failed to upload image",
				"error":   err.Error(),
			})
		}
	}

	// เตรียม field และ values สำหรับ update
	fields := []string{}
	values := []interface{}{}

	if firstname != "" {
		fields = append(fields, "firstname")
		values = append(values, firstname)
	}
	if lastname != "" {
		fields = append(fields, "lastname")
		values = append(values, lastname)
	}
	if email != "" {
		fields = append(fields, "email")
		values = append(values, email)
	}
	if phone != "" {
		fields = append(fields, "phone")
		values = append(values, phone)
	}
	if aboutme != "" {
		fields = append(fields, "aboutme")
		values = append(values, aboutme)
	}
	if imageURL != "" {
		fields = append(fields, "image_url")
		values = append(values, imageURL)
	}

	if len(fields) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "No data to update"})
	}

	// where condition
	whereCon := fmt.Sprintf("user_id = $%d", len(fields)+1)
	values = append(values, userID)

	// อัปเดตข้อมูล
	rowsAffected, err := ac.AuthService.DBService.UpdateData("users", fields, whereCon, values)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to update user profile", "error": err.Error()})
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
	fields := []string{"user_id", "firstname", "lastname", "email", "phone", "aboutme", "image_url"}

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

	fields := []string{"user_id", "firstname", "lastname", "email", "phone", "aboutme", "image_url"}
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
		Firstname: row["firstname"].(string),
		Lastname:  row["lastname"].(string),
		Email:     row["email"].(string),
		Phone:     row["phone"].(string),
		AboutMe:   row["aboutme"].(string),
		ImageURL:  row["image_url"].(string),
	}

	return c.JSON(http.StatusOK, profile)
}
