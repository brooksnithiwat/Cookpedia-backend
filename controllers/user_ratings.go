package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (ac *AuthController) GetRateScore(c echo.Context) error {
	// ดึง user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	// ดึง post ID จาก endpoint parameter
	postID, err := strconv.ParseInt(c.Param("postId"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post ID"})
	}

	// แปลง uid เป็น int64
	uidStr := fmt.Sprintf("%v", uid)
	userIDInt64, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user ID"})
	}

	fieldsUser := []string{"user_id", "post_id", "rate"}

	// SELECT user_id, post_id, rate FROM rating WHERE user_id = $1 AND post_id = $2
	rateData, err := ac.AuthService.DBService.SelectData(
		"rating",
		fieldsUser,
		true,
		"user_id = $1 AND post_id = $2",
		[]interface{}{userIDInt64, postID},
		false,
		"",
		"",
		"",
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get user rate"})
	}

	// ถ้าหาไม่เจอ
	if len(rateData) == 0 {
		return c.JSON(http.StatusOK, echo.Map{
			"rate": nil, // ผู้ใช้ยังไม่ได้ rate
		})
	}

	// return rate
	return c.JSON(http.StatusOK, echo.Map{
		"rate": rateData[0]["rate"],
	})
}

func (ac *AuthController) RatePost(c echo.Context) error {

	// ✅ ดึง user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}
	userID, _ := strconv.ParseInt(fmt.Sprintf("%v", uid), 10, 64)

	// ✅ ดึงค่า post_id และ rate จาก form-data
	postIDStr := c.FormValue("post_id")
	rateStr := c.FormValue("rate")

	if postIDStr == "" || rateStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "post_id and rate are required"})
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil || postID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid post_id"})
	}

	rate, err := strconv.Atoi(rateStr)
	if err != nil || rate < 1 || rate > 5 {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Rate must be between 1 and 5"})
	}

	// ✅ พยายาม INSERT ก่อน
	data := map[string]interface{}{
		"user_id": userID,
		"post_id": postID,
		"rate":    rate,
	}

	_, err = ac.AuthService.DBService.InsertData("rating", data)
	if err != nil {
		// ✅ ถ้า unique violation → UPDATE แทน
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			fields := []string{"rate"}
			condition := "user_id = $2 AND post_id = $3"
			values := []interface{}{rate, userID, postID}

			rowsAffected, err2 := ac.AuthService.DBService.UpdateData("rating", fields, condition, values)
			if err2 != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to update rating", "error": err2.Error()})
			}
			if rowsAffected == 0 {
				return c.JSON(http.StatusNotFound, echo.Map{"message": "Rating not found to update"})
			}

			return c.JSON(http.StatusOK, echo.Map{"message": "Rating updated successfully"})
		}

		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Failed to rate post",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Rating submitted successfully"})
}
func (ac *AuthController) GetRatePost(c echo.Context) error {
	// 1) ตรวจสอบ user_id จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	userIDStr := fmt.Sprintf("%v", uid)
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Invalid user_id"})
	}

	// 2) ดึง post IDs + rate ของ user จาก table rating
	ratedPosts, err := ac.AuthService.DBService.GetRatedPostsByUser(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to fetch rated posts", "error": err.Error()})
	}

	// 3) สำหรับแต่ละ post_id ดึงรายละเอียดโพสต์และ avg rate
	posts := []map[string]interface{}{}
	for _, rp := range ratedPosts {
		postData, err := ac.AuthService.DBService.GetPostWithTagsAndDetails(rp.PostID)
		if err != nil {
			continue // skip ถ้าโพสต์ไม่เจอ
		}

		avgRate, _ := ac.AuthService.DBService.GetAvgRating(rp.PostID) // avg rate ของโพสต์

		post := map[string]interface{}{
			"post_id":   postData["post_id"].(int),
			"image_url": postData["image_url"].(string),
			"star":      avgRate, // avg rate ของโพสต์
		}

		posts = append(posts, post)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Rated posts fetched successfully",
		"posts":   posts,
	})
}
