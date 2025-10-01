package controllers

import (
	"encoding/json"
	"fmt"
	"go-auth/models"
	"go-auth/supabaseutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (ac *AuthController) CreatePost(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	uidStr := fmt.Sprintf("%v", userID)
	userIDInt, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Invalid user_id"})
	}

	// อ่านค่า form
	menuName := c.FormValue("menu_name")
	story := c.FormValue("Details")
	categoriesStr := c.FormValue("categories_tags")       // [1,3]
	ingredientsTagsStr := c.FormValue("ingredients_tags") // [2,5,7]
	ingredientsStr := c.FormValue("ingredients")          // ["Pork","Cooked Rice"]
	instructionsStr := c.FormValue("instructions")        // ["Step1","Step2"]

	if menuName == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "menu_name is required"})
	}

	// แปลง JSON เป็น slice
	var categoryIDs, ingredientTagIDs []int
	err = json.Unmarshal([]byte(categoriesStr), &categoryIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid categories_tags format"})
	}

	err = json.Unmarshal([]byte(ingredientsTagsStr), &ingredientTagIDs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid ingredients_tags format"})
	}

	var ingredients, instructions []string
	err = json.Unmarshal([]byte(ingredientsStr), &ingredients)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid ingredients format"})
	}

	err = json.Unmarshal([]byte(instructionsStr), &instructions)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid instructions format"})
	}

	// อัปโหลดไฟล์ถ้ามี
	var imageURL string
	file, _ := c.FormFile("image")
	if file != nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to open image"})
		}
		defer src.Close()
		imageURL, err = supabaseutil.UploadFile(src, file, userIDInt, "PostImage/post")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to upload image", "error": err.Error()})
		}
	}

	// เริ่ม transaction
	tx, err := ac.AuthService.DBService.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to begin transaction", "error": err.Error()})
	}
	defer tx.Rollback()

	// insert post
	var postID int
	err = tx.QueryRow(
		"INSERT INTO posts (user_id, menu_name, story, image_url) VALUES ($1,$2,$3,$4) RETURNING post_id",
		userIDInt, menuName, story, imageURL,
	).Scan(&postID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert post", "error": err.Error()})
	}

	// insert post_categories
	for _, catID := range categoryIDs {
		_, err := tx.Exec("INSERT INTO post_categories (post_id, category_tag_id) VALUES ($1,$2)", postID, catID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert post_categories", "error": err.Error()})
		}
	}

	// insert post_ingredients (tags)
	for _, ingTagID := range ingredientTagIDs {
		_, err := tx.Exec("INSERT INTO post_ingredients (post_id, ingredient_tag_id) VALUES ($1,$2)", postID, ingTagID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert post_ingredients", "error": err.Error()})
		}
	}

	// insert ingredients_detail
	for _, ing := range ingredients {
		_, err := tx.Exec("INSERT INTO ingredients_detail (post_id, detail) VALUES ($1,$2)", postID, ing)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert ingredients_detail", "error": err.Error()})
		}
	}

	// insert instructions
	for i, step := range instructions {
		_, err := tx.Exec("INSERT INTO instructions (post_id, step_number, detail) VALUES ($1,$2,$3)", postID, i+1, step)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert instructions", "error": err.Error()})
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to commit transaction", "error": err.Error()})
	}

	// ดึงข้อมูลครบพร้อม join
	postData, _ := ac.AuthService.DBService.GetPostWithTagsAndDetails(postID)

	resp := models.PostResponse{
		PostID:          postData["post_id"].(int),
		MenuName:        postData["menu_name"].(string),
		Story:           postData["story"].(string),
		ImageURL:        postData["image_url"].(string),
		CategoriesTags:  postData["categories_tags"].([]string),
		IngredientsTags: postData["ingredients_tags"].([]string),
		Ingredients:     postData["ingredients"].([]string),
		Instructions:    postData["instructions"].([]string),
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"message": "Post created successfully",
		"post":    resp,
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
