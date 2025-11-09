package controllers

import (
	"encoding/json"
	"fmt"
	"go-auth/models"
	"go-auth/supabaseutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func (ac *AuthController) SeachPost(c echo.Context) error {
	// Accept path param /searchpost/:name or query param ?q=...
	name := c.Param("name")
	if strings.TrimSpace(name) == "" {
		name = c.QueryParam("q")
	}
	if strings.TrimSpace(name) == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing search term (path param :name or query param q)"})
	}

	// find matching post IDs (case-insensitive)
	rows, err := ac.AuthService.DBService.DB.Query("SELECT post_id FROM posts WHERE menu_name ILIKE $1 ORDER BY post_id DESC", "%"+name+"%")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to search posts", "error": err.Error()})
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			fmt.Printf("[DEBUG] SeachPost: scan post_id error: %v\n", err)
			continue
		}
		postIDs = append(postIDs, id)
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed reading search results", "error": err.Error()})
	}

	posts := []models.PostResponse{}
	for _, pid := range postIDs {
		postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(pid)
		if err != nil {
			fmt.Printf("[DEBUG] SeachPost: failed to load post %d: %v\n", pid, err)
			continue
		}

		uid := postData["user_id"]
		fields := []string{"user_id", "username", "image_url"}
		whereCon := "user_id = ?"
		whereArgs := []interface{}{uid}
		users, err := ac.AuthService.DBService.SelectData("users", fields, true, whereCon, whereArgs, false, "", "", "")
		if err != nil || len(users) == 0 {
			fmt.Printf("[DEBUG] SeachPost: failed to load user for post %d: err=%v user_count=%d\n", pid, err, len(users))
			continue
		}
		userData := users[0]

		createdAt := ac.AuthService.DBService.ParseDateTime(postData["created_at"], "Asia/Bangkok")

		owner := models.OwnerPost{
			ProfileImage: fmt.Sprintf("%v", userData["profile_image"]),
			Username:     fmt.Sprintf("%v", userData["username"]),
			CreatedDate:  createdAt.Format("2006-01-02"),
			CreatedTime:  createdAt.Format("15:04:05"),
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
		posts = append(posts, resp)
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Search results", "posts": posts, "post_count": len(posts)})
}

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

//		return c.JSON(http.StatusOK, echo.Map{"message": "Post deleted successfully"})
//	}

func (ac *AuthController) GetAllPost(c echo.Context) error {

	// 1) ดึง post_ids ทั้งหมด
	postIDs, err := ac.AuthService.DBService.GetAllPostIDs()
	if err != nil {
		fmt.Println("[ERROR] GetAllPostIDs error:", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to fetch posts"})
	}

	fmt.Println("[DEBUG] All PostIDs:", postIDs)

	posts := []models.PostResponse{}

	// 2) Loop ทุก postID
	for _, postID := range postIDs {

		fmt.Println("[DEBUG] Fetching postID:", postID)

		// ✅ ดึงทุกข้อมูลของ post (tags, ingredients, instructions)
		postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(postID)
		fmt.Println("[DEBUG] Result for postID =", postID, "| postData =", postData, "| err =", err)

		if err != nil {
			fmt.Println("[ERROR] GetPostWithTagsAndDetails error for post", postID, ":", err)
			continue
		}

		// ✅ ต้องมี user_id ไม่งั้นข้าม
		ownerID, ok := postData["user_id"].(int)
		if !ok {
			fmt.Println("[ERROR] ownerID missing for post", postID)
			continue
		}

		fmt.Println("[DEBUG] ownerID for post", postID, "=", ownerID)

		// ✅ ดึงข้อมูลเจ้าของโพสต์
		fields := []string{"user_id", "username", "image_url"}
		users, err := ac.AuthService.DBService.SelectData(
			"users",
			fields,
			true,
			"user_id = ?",
			[]interface{}{ownerID},
			false,
			"",
			"",
			"",
		)

		fmt.Println("[DEBUG] SelectData user:", users, "| err =", err)

		if err != nil || len(users) == 0 {
			fmt.Println("[ERROR] Cannot fetch user info for user_id =", ownerID)
			continue
		}

		userData := users[0]

		// ✅ created_at จำเป็นต้องมีใน postData
		createdAt := ac.AuthService.DBService.ParseDateTime(postData["created_at"], "Asia/Bangkok")

		// ✅ Owner struct
		owner := models.OwnerPost{
			ProfileImage: fmt.Sprintf("%v", userData["profile_image"]),
			Username:     fmt.Sprintf("%v", userData["username"]),
			CreatedDate:  createdAt.Format("2006-01-02"),
			CreatedTime:  createdAt.Format("15:04:05"),
		}

		// ✅ PostDetail struct
		post := models.PostDetail{
			PostID:          postData["post_id"].(int),
			MenuName:        fmt.Sprintf("%v", postData["menu_name"]),
			Story:           fmt.Sprintf("%v", postData["story"]),
			ImageURL:        fmt.Sprintf("%v", postData["image_url"]),
			CategoriesTags:  postData["categories_tags"].([]string),
			IngredientsTags: postData["ingredients_tags"].([]string),
			Ingredients:     postData["ingredients"].([]string),
			Instructions:    postData["instructions"].([]string),
		}

		// ✅ รวมข้อมูลลงใน response list
		posts = append(posts, models.PostResponse{
			OwnerPost: owner,
			Post:      post,
		})
	}

	fmt.Println("[DEBUG] Final posts length:", len(posts))

	return c.JSON(http.StatusOK, echo.Map{
		"message": "All posts fetched successfully",
		"posts":   posts,
	})
}

func (ac *AuthController) GetAllMyPost(c echo.Context) error {
	// 1) ตรวจสอบ user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	uidStr := fmt.Sprintf("%v", uid)
	userIDInt64, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Invalid user_id"})
	}

	// 2) ดึง post IDs ของ user คนนี้
	postIDs, err := ac.AuthService.DBService.GetPostIDsByUser(userIDInt64)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to fetch posts", "error": err.Error()})
	}

	// 3) ดึงข้อมูล user (owner) ครั้งเดียว
	fields := []string{"user_id", "username", "image_url"}
	whereCon := "user_id = ?"
	whereArgs := []interface{}{userIDInt64}
	users, err := ac.AuthService.DBService.SelectData(
		"users",
		fields,
		true,
		whereCon,
		whereArgs,
		false,
		"",
		"",
		"",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get user", "error": err.Error()})
	}
	if len(users) == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	}
	userData := users[0]

	// 4) สำหรับแต่ละ post_id ดึงรายละเอียดครบ
	// Return a minimal response shape (owner_post + post with only post_id and image_url)
	posts := []map[string]interface{}{}
	for _, pid := range postIDs {
		postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(pid)
		if err != nil {
			// skip problematic posts
			continue
		}

		// createdAt may be nil in postData; ParseDateTime handles nil
		createdAt := ac.AuthService.DBService.ParseDateTime(postData["created_at"], "Asia/Bangkok")

		owner := map[string]interface{}{
			"profile_image": fmt.Sprintf("%v", userData["profile_image"]),
			"username":      fmt.Sprintf("%v", userData["username"]),
			"created_date":  createdAt.Format("2006-01-02"),
			"created_time":  createdAt.Format("15:04:05"),
		}

		post := map[string]interface{}{
			"post_id":   postData["post_id"].(int),
			"image_url": postData["image_url"].(string),
		}

		resp := map[string]interface{}{
			"owner_post": owner,
			"post":       post,
		}

		posts = append(posts, resp)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "All posts fetched successfully",
		"posts":   posts,
	})
}

// GetAllPostByUsername returns all posts created by the given username (public)
func (ac *AuthController) GetAllPostByUsername(c echo.Context) error {
	username := c.Param("username")
	if strings.TrimSpace(username) == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing username in path"})
	}

	// Find user_id by username
	fields := []string{"user_id", "username", "image_url"}
	whereCon := "username = ?"
	whereArgs := []interface{}{username}
	users, err := ac.AuthService.DBService.SelectData("users", fields, true, whereCon, whereArgs, false, "", "", "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to lookup user", "error": err.Error()})
	}
	if len(users) == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	}
	userData := users[0]

	// user_id may be int or int64 or string; normalize to int64
	var userIDInt64 int64
	switch v := userData["user_id"].(type) {
	case int64:
		userIDInt64 = v
	case int:
		userIDInt64 = int64(v)
	case float64:
		userIDInt64 = int64(v)
	case string:
		tmp, _ := strconv.ParseInt(v, 10, 64)
		userIDInt64 = tmp
	default:
		userIDInt64 = 0
	}

	// Get post IDs by user
	postIDs, err := ac.AuthService.DBService.GetPostIDsByUser(userIDInt64)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to fetch posts", "error": err.Error()})
	}

	posts := []models.PostResponse{}
	for _, pid := range postIDs {
		postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(pid)
		if err != nil {
			fmt.Printf("[DEBUG] GetAllPostByUsername: failed to load post %d: %v\n", pid, err)
			continue
		}

		createdAt := ac.AuthService.DBService.ParseDateTime(postData["created_at"], "Asia/Bangkok")

		owner := models.OwnerPost{
			ProfileImage: fmt.Sprintf("%v", userData["profile_image"]),
			Username:     fmt.Sprintf("%v", userData["username"]),
			CreatedDate:  createdAt.Format("2006-01-02"),
			CreatedTime:  createdAt.Format("15:04:05"),
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
		posts = append(posts, resp)
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "User posts fetched", "posts": posts, "post_count": len(posts)})
}

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

//		return c.JSON(http.StatusOK, echo.Map{
//			"post": resp,
//		})
//	}
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
	categoriesStr := c.FormValue("categories_tags")
	ingredientsTagsStr := c.FormValue("ingredients_tags")
	ingredientsStr := c.FormValue("ingredients")
	instructionsStr := c.FormValue("instructions")

	// Validate input
	if menuName == "" || story == "" || categoriesStr == "" || ingredientsTagsStr == "" || ingredientsStr == "" || instructionsStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Missing required fields"})
	}

	// Convert JSON input to structs
	var categoryIDs, ingredientTagIDs []int
	if err := json.Unmarshal([]byte(categoriesStr), &categoryIDs); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid categories_tags format"})
	}
	if err := json.Unmarshal([]byte(ingredientsTagsStr), &ingredientTagIDs); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid ingredients_tags format"})
	}

	var ingredients, instructions []string
	if err := json.Unmarshal([]byte(ingredientsStr), &ingredients); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid ingredients format"})
	}
	if err := json.Unmarshal([]byte(instructionsStr), &instructions); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid instructions format"})
	}

	// Image upload
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

	// Start transaction
	tx, err := ac.AuthService.DBService.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to begin transaction", "error": err.Error()})
	}
	defer tx.Rollback()

	// Insert post
	var postID int
	err = tx.QueryRow(
		"INSERT INTO posts (user_id, menu_name, story, image_url) VALUES ($1,$2,$3,$4) RETURNING post_id",
		userIDInt, menuName, story, imageURL,
	).Scan(&postID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert post", "error": err.Error()})
	}

	// Insert category tags
	for _, catID := range categoryIDs {
		_, err := tx.Exec("INSERT INTO post_categories (post_id, category_tag_id) VALUES ($1,$2)", postID, catID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert post_categories", "error": err.Error()})
		}
	}

	// Insert ingredient tags
	for _, ingTagID := range ingredientTagIDs {
		_, err := tx.Exec("INSERT INTO post_ingredients (post_id, ingredient_tag_id) VALUES ($1,$2)", postID, ingTagID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert post_ingredients", "error": err.Error()})
		}
	}

	// Insert ingredient details
	for _, ing := range ingredients {
		_, err := tx.Exec("INSERT INTO ingredients_detail (post_id, detail) VALUES ($1,$2)", postID, ing)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert ingredients_detail", "error": err.Error()})
		}
	}

	// Insert instructions
	for i, step := range instructions {
		_, err := tx.Exec("INSERT INTO instructions (post_id, step_number, detail) VALUES ($1,$2,$3)", postID, i+1, step)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert instructions", "error": err.Error()})
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to commit transaction", "error": err.Error()})
	}

	// ดึงข้อมูล post พร้อม tags, ingredients, instructions
	postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(postID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get post details", "error": err.Error()})
	}

	// Safe type assertions
	postIDVal, ok := postData["post_id"]
	if !ok || postIDVal == nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Post ID is nil"})
	}
	var postIDIntSafe int
	switch v := postIDVal.(type) {
	case int:
		postIDIntSafe = v
	case int64:
		postIDIntSafe = int(v)
	default:
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Post ID type is invalid"})
	}

	// CategoriesTags
	cats, ok := postData["categories_tags"].([]string)
	if !ok || cats == nil {
		cats = []string{}
	}

	// IngredientsTags
	ingTags, ok := postData["ingredients_tags"].([]string)
	if !ok || ingTags == nil {
		ingTags = []string{}
	}

	// Ingredients
	ings, ok := postData["ingredients"].([]string)
	if !ok || ings == nil {
		ings = []string{}
	}

	// Instructions
	ins, ok := postData["instructions"].([]string)
	if !ok || ins == nil {
		ins = []string{}
	}

	// created_at
	createdAtVal := postData["created_at"]
	var createdAt time.Time
	if createdAtVal != nil {
		createdAt = ac.AuthService.DBService.ParseDateTime(createdAtVal, "Asia/Bangkok")
	} else {
		createdAt = time.Now()
	}

	// ดึง user data
	fields := []string{"user_id", "username", "image_url"}
	whereCon := "user_id = ?"
	whereArgs := []interface{}{userIDInt}

	users, err := ac.AuthService.DBService.SelectData(
		"users",
		fields,
		true,
		whereCon,
		whereArgs,
		false,
		"",
		"",
		"",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get user", "error": err.Error()})
	}
	if len(users) == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	}
	userData := users[0]

	owner := models.OwnerPost{
		ProfileImage: fmt.Sprintf("%v", userData["profile_image"]),
		Username:     fmt.Sprintf("%v", userData["username"]),
		CreatedDate:  createdAt.Format("2006-01-02"),
		CreatedTime:  createdAt.Format("15:04:05"),
	}

	post := models.PostDetail{
		PostID:          postIDIntSafe,
		MenuName:        fmt.Sprintf("%v", postData["menu_name"]),
		Story:           fmt.Sprintf("%v", postData["story"]),
		ImageURL:        fmt.Sprintf("%v", postData["image_url"]),
		CategoriesTags:  cats,
		IngredientsTags: ingTags,
		Ingredients:     ings,
		Instructions:    ins,
	}

	resp := models.PostResponse{
		OwnerPost: owner,
		Post:      post,
	}

	return c.JSON(http.StatusOK, resp)
}
