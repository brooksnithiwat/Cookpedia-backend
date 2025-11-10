package controllers

import (
	"fmt"
	"go-auth/models"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (ac *AuthController) GetCommentsByPostID(c echo.Context) error {
	// 1) ดึง post_id จาก param
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post_id"})
	}

	// 2) ดึง comment จาก DB โดยใช้ SelectData
	fields := []string{
		"c.comment_id",
		"c.content",
		"c.created_at",
		"u.user_id",
		"u.username",
		"u.image_url",
	}
	commentsData, err := ac.AuthService.DBService.SelectData(
		"comment c",                 // table
		fields,                      // fields
		true,                        // where
		"c.post_id = ?",             // where condition
		[]interface{}{postID},       // where values
		true,                        // join
		"users u",                   // join table
		"c.user_id = u.user_id",     // join condition
		"ORDER BY c.created_at ASC", // order
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to fetch comments",
			"error":   err.Error(),
		})
	}

	comments := []models.CommentResponse{}

	for _, row := range commentsData {
		createdAt := ac.AuthService.DBService.ParseDateTime(row["created_at"], "Asia/Bangkok")

		comment := models.CommentResponse{
			CommentID:  row["comment_id"].(int64),
			Content:    fmt.Sprintf("%v", row["content"]),
			CreatedAt:  createdAt.Format("2006-01-02 15:04:05"),
			UserID:     row["user_id"].(int64),
			Username:   fmt.Sprintf("%v", row["username"]),
			ProfileImg: fmt.Sprintf("%v", row["profile_image"]),
		}

		comments = append(comments, comment)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message":  "Comments fetched successfully",
		"comments": comments,
	})
}

func (ac *AuthController) AddComment(c echo.Context) error {
	// ดึง user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}
	userID, _ := strconv.ParseInt(fmt.Sprintf("%v", uid), 10, 64)

	// ดึง post_id และ content จาก form-data
	postIDStr := c.FormValue("post_id")
	content := c.FormValue("content")

	if content == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Content cannot be empty"})
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post_id"})
	}

	// ตรวจสอบว่าโพสต์มีอยู่
	exists, err := ac.AuthService.DBService.PostExists(int(postID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to check post", "error": err.Error()})
	}
	if !exists {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Post not found"})
	}

	// เพิ่ม comment
	data := map[string]interface{}{
		"user_id": userID,
		"post_id": postID,
		"content": content,
	}

	_, err = ac.AuthService.DBService.InsertData("comment", data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to add comment", "error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Comment added successfully"})
}

func (ac *AuthController) EditComment(c echo.Context) error {
	// 1) ตรวจสอบ user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}
	userID, _ := strconv.ParseInt(fmt.Sprintf("%v", uid), 10, 64)

	// 2) รับ form-data
	commentIDStr := c.FormValue("comment_id")
	newContent := c.FormValue("content")

	if newContent == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Content cannot be empty"})
	}

	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid comment_id"})
	}

	// 3) ตรวจสอบว่า comment เป็นของ user นี้หรือไม่
	exists, err := ac.AuthService.DBService.CommentBelongsToUser(int(commentID), int(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to check comment ownership", "error": err.Error()})
	}
	if !exists {
		return c.JSON(http.StatusForbidden, echo.Map{"message": "You did not have a permission to edit this comment"})
	}

	// 4) อัปเดต comment
	fields := []string{"content"}
	condition := "comment_id = $2"
	values := []interface{}{newContent, commentID}

	rowsAffected, err := ac.AuthService.DBService.UpdateData("comment", fields, condition, values)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to update comment", "error": err.Error()})
	}
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Comment not found"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Comment updated successfully"})
}
func (ac *AuthController) DeleteCommentByCommentID(c echo.Context) error {
	// 1) ดึง user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}
	userID, _ := strconv.ParseInt(fmt.Sprintf("%v", uid), 10, 64)

	// 2) ดึง comment_id จาก param
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid comment_id"})
	}

	// 3) ตรวจสอบว่า comment เป็นของ user คนนี้
	exists, err := ac.AuthService.DBService.SelectData(
		"comment",
		[]string{"comment_id"},
		true,
		"comment_id = ? AND user_id = ?",
		[]interface{}{commentID, userID},
		false,
		"", "", "",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to check comment ownership", "error": err.Error()})
	}
	if len(exists) == 0 {
		return c.JSON(http.StatusForbidden, echo.Map{"message": "You are not the owner of this comment"})
	}

	// 4) ลบ comment
	rowsAffected, err := ac.AuthService.DBService.DeleteData(
		"comment",
		"comment_id = $1",
		[]interface{}{commentID},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to delete comment", "error": err.Error()})
	}
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Comment not found"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Comment deleted successfully"})
}
