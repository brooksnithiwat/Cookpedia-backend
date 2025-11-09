package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (ac *AuthController) FavoritePost(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	postID, err := strconv.ParseInt(c.Param("postId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post ID"})
	}

	data := map[string]interface{}{
		"user_id": userID,
		"post_id": postID,
	}

	// Insert using InsertData
	_, err = ac.AuthService.DBService.InsertData("favorite", data)
	if err != nil {
		// ถ้าเป็น duplicate
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "Unique") {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Post already favorited"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to favorite post"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Post favorited successfully"})
}

func (ac *AuthController) UnFavoritePost(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	postID, err := strconv.ParseInt(c.Param("postId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post ID"})
	}

	// ใช้ DeleteData ของ DatabaseService
	condition := "user_id = $1 AND post_id = $2"
	conditionValues := []interface{}{userID, postID}

	rowsAffected, err := ac.AuthService.DBService.DeleteData("favorite", condition, conditionValues)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to unfavorite post"})
	}

	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Favorite post not found"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Post unfavorited successfully"})
}
func (ac *AuthController) GetAllFavoritePost(c echo.Context) error {
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

	// 2) ดึง post IDs ที่ user กด favorite
	fields := []string{"post_id"}
	whereCon := "user_id = $1"
	whereArgs := []interface{}{userIDInt64}

	favPosts, err := ac.AuthService.DBService.SelectData(
		"favorite",
		fields,
		true,
		whereCon,
		whereArgs,
		false,
		"", "", "",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to fetch favorite posts", "error": err.Error()})
	}

	// ถ้าไม่มี favorite
	if len(favPosts) == 0 {
		return c.JSON(http.StatusOK, echo.Map{"message": "No favorite posts", "posts": []interface{}{}})
	}

	// 3) ดึงข้อมูล user (owner) ครั้งเดียว
	fieldsUser := []string{"user_id", "username", "image_url"}
	users, err := ac.AuthService.DBService.SelectData(
		"users",
		fieldsUser,
		true,
		"user_id = $1",
		[]interface{}{userIDInt64},
		false,
		"", "", "",
	)
	if err != nil || len(users) == 0 {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get user"})
	}
	userData := users[0]

	// 4) Loop ทุก post_id ดึง post detail
	posts := []map[string]interface{}{}
	for _, row := range favPosts {
		// post_id can be int, int64, float64, or string depending on driver; handle safely
		var postID int
		switch v := row["post_id"].(type) {
		case int:
			postID = v
		case int64:
			postID = int(v)
		case float64:
			postID = int(v)
		case string:
			tmp, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			postID = tmp
		default:
			continue
		}

		postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(postID)
		if err != nil {
			continue
		}

		createdAt := ac.AuthService.DBService.ParseDateTime(postData["created_at"], "Asia/Bangkok")

		owner := map[string]interface{}{
			"profile_image": fmt.Sprintf("%v", userData["profile_image"]),
			"username":      fmt.Sprintf("%v", userData["username"]),
			"created_date":  createdAt.Format("2006-01-02"),
			"created_time":  createdAt.Format("15:04:05"),
		}

		post := map[string]interface{}{
			"post_id":   postData["post_id"].(int),
			"image_url": fmt.Sprintf("%v", postData["image_url"]),
		}

		resp := map[string]interface{}{
			"owner_post": owner,
			"post":       post,
		}

		posts = append(posts, resp)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Favorite posts fetched successfully",
		"posts":   posts,
	})
}
