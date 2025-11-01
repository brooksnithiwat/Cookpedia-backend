package controllers

import (
	"encoding/json"
	"fmt"
	"go-auth/models"
	"go-auth/supabaseutil"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// func (ac *AuthController) GetAllPost(c echo.Context) error {
// 	// 1) ดึง post ทั้งหมด
// 	postIDs, err := ac.AuthService.DBService.GetAllPostIDs()
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to fetch posts"})
// 	}

// 	posts := []models.PostResponse{}

// 	// 2) สำหรับแต่ละ post_id ดึงรายละเอียดครบ
// 	for _, postID := range postIDs {
// 		postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(postID)
// 		if err != nil {
// 			continue // ถ้า post นี้มีปัญหา skip ไป
// 		}

// 		post := models.PostResponse{
// 			PostID:          postData["post_id"].(int),
// 			MenuName:        postData["menu_name"].(string),
// 			Story:           postData["story"].(string),
// 			ImageURL:        postData["image_url"].(string),
// 			CategoriesTags:  postData["categories_tags"].([]string),
// 			IngredientsTags: postData["ingredients_tags"].([]string),
// 			Ingredients:     postData["ingredients"].([]string),
// 			Instructions:    postData["instructions"].([]string),
// 		}

// 		posts = append(posts, post)
// 	}

// 	return c.JSON(http.StatusOK, echo.Map{
// 		"message": "All posts fetched successfully",
// 		"posts":   posts,
// 	})
// }

// func (ac *AuthController) DeletePostbyPostID(c echo.Context) error {
// 	// 1) ตรวจสอบ user_id จาก token
// 	uid := c.Get("user_id")
// 	if uid == nil {
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
// 	}
// 	userID, _ := strconv.ParseInt(fmt.Sprintf("%v", uid), 10, 64)

// 	// 2) ดึง post_id จาก path param
// 	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post_id"})
// 	}

// 	// 3) ตรวจสอบว่า user เป็นเจ้าของโพสต์
// 	var exists bool
// 	err = ac.AuthService.DBService.DB.QueryRow(
// 		"SELECT EXISTS(SELECT 1 FROM posts WHERE post_id=$1 AND user_id=$2)",
// 		postID, userID,
// 	).Scan(&exists)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to check post ownership", "error": err.Error()})
// 	}
// 	if !exists {
// 		return c.JSON(http.StatusForbidden, echo.Map{"message": "You don't have permission to delete this post"})
// 	}

// 	// 4) ลบข้อมูลจาก table ที่เกี่ยวข้อง
// 	tables := []string{"post_categories", "post_ingredients", "ingredients_detail", "instructions"}
// 	for _, table := range tables {
// 		rowsAffected, err := ac.AuthService.DBService.DeleteData(table, "post_id=$1", []interface{}{postID})
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"message": fmt.Sprintf("Failed to delete from %s", table), "error": err.Error()})
// 		}
// 		fmt.Printf("[DEBUG] Deleted %d rows from %s\n", rowsAffected, table)
// 	}

// 	// 5) ลบโพสต์หลัก
// 	rowsAffected, err := ac.AuthService.DBService.DeleteData("posts", "post_id=$1", []interface{}{postID})
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to delete post", "error": err.Error()})
// 	}
// 	if rowsAffected == 0 {
// 		return c.JSON(http.StatusNotFound, echo.Map{"message": "Post not found"})
// 	}

// 	return c.JSON(http.StatusOK, echo.Map{"message": "Post deleted successfully"})
// }

// func (ac *AuthController) GetAllMyPost(c echo.Context) error {
// 	// 1) ตรวจสอบ user_id จาก token
// 	userID := c.Get("user_id")
// 	if userID == nil {
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
// 	}

// 	uidStr := fmt.Sprintf("%v", userID)
// 	userIDInt, err := strconv.ParseInt(uidStr, 10, 64)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Invalid user_id"})
// 	}

// 	// 2) ดึงโพสต์ทั้งหมดของ user
// 	postIDs, err := ac.AuthService.DBService.GetPostIDsByUser(userIDInt)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to fetch posts"})
// 	}

// 	posts := []models.PostResponse{}

// 	// 3) สำหรับแต่ละ post_id ดึงรายละเอียดครบ
// 	for _, postID := range postIDs {
// 		postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(postID)
// 		if err != nil {
// 			continue // ถ้า post นี้มีปัญหา skip ไป
// 		}

// 		post := models.PostResponse{
// 			PostID:          postData["post_id"].(int),
// 			MenuName:        postData["menu_name"].(string),
// 			Story:           postData["story"].(string),
// 			ImageURL:        postData["image_url"].(string),
// 			CategoriesTags:  postData["categories_tags"].([]string),
// 			IngredientsTags: postData["ingredients_tags"].([]string),
// 			Ingredients:     postData["ingredients"].([]string),
// 			Instructions:    postData["instructions"].([]string),
// 		}

// 		posts = append(posts, post)
// 	}

// 	return c.JSON(http.StatusOK, echo.Map{
// 		"message": "All posts fetched successfully",
// 		"posts":   posts,
// 	})
// }

// func (ac *AuthController) EditPostByPostID(c echo.Context) error {
// 	// 1) ตรวจสอบ user_id จาก token
// 	uid := c.Get("user_id")
// 	if uid == nil {
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
// 	}
// 	userID, _ := strconv.ParseInt(fmt.Sprintf("%v", uid), 10, 64)

// 	// 2) ดึง post_id จาก path param
// 	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post_id"})
// 	}

// 	// 3) ดึง form params
// 	formParams, err := c.FormParams()
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid form data"})
// 	}

// 	fields := []string{}
// 	values := []interface{}{}

// 	// ตรวจสอบ field หลัก
// 	if _, ok := formParams["menu_name"]; ok {
// 		fields = append(fields, "menu_name")
// 		values = append(values, c.FormValue("menu_name"))
// 	}
// 	if _, ok := formParams["story"]; ok {
// 		fields = append(fields, "story")
// 		values = append(values, c.FormValue("story"))
// 	}

// 	// รูปภาพ
// 	file, _ := c.FormFile("image")
// 	if file != nil {
// 		src, err := file.Open()
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to open image"})
// 		}
// 		defer src.Close()
// 		imageURL, err := supabaseutil.UploadFile(src, file, userID, "PostImage/post")
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to upload image", "error": err.Error()})
// 		}
// 		fields = append(fields, "image_url")
// 		values = append(values, imageURL)
// 	}

// 	// ถ้าไม่มี field หลัก + รูป ส่งมาเลย
// 	if len(fields) == 0 && file == nil && len(formParams) == 0 {
// 		return c.JSON(http.StatusBadRequest, echo.Map{"message": "No data to update"})
// 	}

// 	// 4) เริ่ม transaction
// 	tx, err := ac.AuthService.DBService.DB.Begin()
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to begin transaction", "error": err.Error()})
// 	}
// 	defer tx.Rollback()

// 	// 5) update posts table
// 	if len(fields) > 0 {
// 		whereCon := fmt.Sprintf("post_id = $%d AND user_id = $%d", len(fields)+1, len(fields)+2)
// 		valuesWithID := append(values, postID, userID)
// 		rowsAffected, err := ac.AuthService.DBService.UpdateData("posts", fields, whereCon, valuesWithID)
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to update post", "error": err.Error()})
// 		}
// 		if rowsAffected == 0 {
// 			return c.JSON(http.StatusNotFound, echo.Map{"message": "Post not found or you don't have permission"})
// 		}
// 	}

// 	// 6) Update categories_tags
// 	if val, ok := formParams["categories_tags"]; ok {
// 		var categoryIDs []int
// 		err := json.Unmarshal([]byte(val[0]), &categoryIDs)
// 		if err != nil {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid categories_tags format"})
// 		}
// 		_, _ = tx.Exec("DELETE FROM post_categories WHERE post_id = $1", postID)
// 		for _, catID := range categoryIDs {
// 			_, _ = tx.Exec("INSERT INTO post_categories (post_id, category_tag_id) VALUES ($1,$2)", postID, catID)
// 		}
// 	}

// 	// 7) Update ingredients_tags
// 	if val, ok := formParams["ingredients_tags"]; ok {
// 		var ingTagIDs []int
// 		err := json.Unmarshal([]byte(val[0]), &ingTagIDs)
// 		if err != nil {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid ingredients_tags format"})
// 		}
// 		_, _ = tx.Exec("DELETE FROM post_ingredients WHERE post_id = $1", postID)
// 		for _, ingID := range ingTagIDs {
// 			_, _ = tx.Exec("INSERT INTO post_ingredients (post_id, ingredient_tag_id) VALUES ($1,$2)", postID, ingID)
// 		}
// 	}

// 	// 8) Update ingredients details
// 	if val, ok := formParams["ingredients"]; ok {
// 		var ingredients []string
// 		err := json.Unmarshal([]byte(val[0]), &ingredients)
// 		if err != nil {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid ingredients format"})
// 		}
// 		_, _ = tx.Exec("DELETE FROM ingredients_detail WHERE post_id = $1", postID)
// 		for _, ing := range ingredients {
// 			_, _ = tx.Exec("INSERT INTO ingredients_detail (post_id, detail) VALUES ($1,$2)", postID, ing)
// 		}
// 	}

// 	// 9) Update instructions
// 	if val, ok := formParams["instructions"]; ok {
// 		var instructions []string
// 		err := json.Unmarshal([]byte(val[0]), &instructions)
// 		if err != nil {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid instructions format"})
// 		}
// 		_, _ = tx.Exec("DELETE FROM instructions WHERE post_id = $1", postID)
// 		for i, step := range instructions {
// 			_, _ = tx.Exec("INSERT INTO instructions (post_id, step_number, detail) VALUES ($1,$2,$3)", postID, i+1, step)
// 		}
// 	}

// 	// commit transaction
// 	if err := tx.Commit(); err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to commit transaction", "error": err.Error()})
// 	}

// 	return c.JSON(http.StatusOK, echo.Map{"message": "Post updated successfully"})
// }

// func (ac *AuthController) GetPostByPostID(c echo.Context) error {
// 	// อ่าน post_id จาก params
// 	postIDStr := c.Param("id")
// 	postID, err := strconv.Atoi(postIDStr)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post id"})
// 	}

// 	// ดึงข้อมูลจาก DB
// 	postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(postID)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, echo.Map{
// 			"message": "Failed to fetch post",
// 			"error":   err.Error(),
// 		})
// 	}

// 	if postData == nil {
// 		return c.JSON(http.StatusNotFound, echo.Map{"message": "Post not found"})
// 	}

// 	// Map ข้อมูลใส่ struct response
// 	resp := models.PostResponse{
// 		PostID:          postData["post_id"].(int),
// 		MenuName:        postData["menu_name"].(string),
// 		Story:           postData["story"].(string),
// 		ImageURL:        postData["image_url"].(string),
// 		CategoriesTags:  postData["categories_tags"].([]string),
// 		IngredientsTags: postData["ingredients_tags"].([]string),
// 		Ingredients:     postData["ingredients"].([]string),
// 		Instructions:    postData["instructions"].([]string),
// 	}

// 	return c.JSON(http.StatusOK, echo.Map{
// 		"post": resp,
// 	})
// }

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

	//ดักว่าต้องส่งทุก field
	if menuName == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing menu name"})
	}
	if story == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing details name"})
	}
	if categoriesStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing category tags"})
	}
	if ingredientsTagsStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing ingredients tags"})
	}
	if ingredientsStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing ingredients detail"})
	}
	if instructionsStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing menu instructions"})
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

	//ดึงค่า date-time ปัจจุบันจากใน post-services
	createdAt := ac.AuthService.DBService.ParseDateTime(postData["created_at"], "Asia/Bangkok")

	fields := []string{"user_id", "username", "image_url"}
	whereCon := "user_id = ?"
	whereArgs := []interface{}{userID}
	users, err := ac.AuthService.DBService.SelectData(
		"users",
		fields,
		true, // where
		whereCon,
		whereArgs,
		false, // join
		"",    // joinTable
		"",    // joinCondition
		"",    // orderAndLimit
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get user", "error": err.Error()})
	}

	if len(users) == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	}

	userData := users[0] // map[string]interface{}

	owner := models.OwnerPost{
		ProfileImage: fmt.Sprintf("%v", userData["profile_image"]),
		Username:     fmt.Sprintf("%v", userData["username"]),
		CreatedDate: createdAt.Format("2006-01-02"),
	    CreatedTime: createdAt.Format("15:04:05"),
	}

	post := models.PostDetail{
		PostID:          postData["post_id"].(int),
		MenuName:        postData["menu_name"].(string),
		Story:           postData["story"].(string),
		ImageURL:        postData["image_url"].(string),
		CategoriesTags:  postData["categories_tags"].([]string),
		IngredientsTags: postData["ingredients_tags"].([]string),
		Ingredients:     postData["ingredients"].([]string),
		Instructions:    postData["instructions"].([]string),
	}

	resp := models.PostResponse{
		OwnerPost: owner,
		Post:      post,
	}

	return c.JSON(http.StatusOK, resp)

}
