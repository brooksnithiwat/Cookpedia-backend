package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)
func (ac *AuthController) GetAllFollowingDetail(c echo.Context) error {
	// 1) ตรวจสอบ token (protected)
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}

	// 2) ดึง user_id จาก param (ใครที่เราต้องการดูว่า follow ใครบ้าง)
	userStr := c.Param("id")
	userID, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user_id"})
	}

	// 3) Query ข้อมูล following
	rows, err := ac.AuthService.DBService.SelectData(
		"followers f INNER JOIN users u ON f.following_id = u.user_id",
		[]string{"u.user_id", "u.firstname", "u.lastname", "u.image_url"},
		true,
		"f.follower_id = $1",
		[]interface{}{userID},
		false,
		"", "", "",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			echo.Map{"message": "Failed to get following details", "error": err.Error()})
	}

	// 4) Return JSON
	return c.JSON(http.StatusOK, echo.Map{
		"following": rows, // []map[string]interface{}
	})
}

func (ac *AuthController) GetAllFollowerDetail(c echo.Context) error {
	//เช็ค protected route จาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}
	// ดึง user_id จาก param
	userStr := c.Param("id")
	userID, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user_id"})
	}
	// 2) Query ข้อมูล follower
	// ดึง follower_id จาก table followers แล้ว join กับ users
	rows, err := ac.AuthService.DBService.SelectData(
		"followers f INNER JOIN users u ON f.follower_id = u.user_id",
		[]string{"u.user_id", "u.firstname", "u.lastname", "u.image_url"},
		true,
		"f.following_id = $1",
		[]interface{}{userID},
		false,
		"", "", "",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			echo.Map{"message": "Failed to get follower details", "error": err.Error()})
	}

	// 3) Return JSON
	return c.JSON(http.StatusOK, echo.Map{
		"followers": rows, // []map[string]interface{} ของ follower
	})
}

func (ac *AuthController) GetAllFollower(c echo.Context) error {
	// 1) ดึง user_id จาก param
	userStr := c.Param("id")
	userID, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user_id"})
	}

	// 2) Query จำนวน follower
	rows, err := ac.AuthService.DBService.SelectData(
		"followers",                   // table
		[]string{"COUNT(*) AS total"}, // fields
		true,                          // where
		"following_id = $1",           // where condition
		[]interface{}{userID},         // args
		false, "", "", "",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			echo.Map{"message": "Failed to get follower count", "error": err.Error()})
	}

	// 3) อ่านผลลัพธ์
	var count int64 = 0
	if len(rows) > 0 {
		if val, ok := rows[0]["total"].(int64); ok {
			count = val
		} else if val, ok := rows[0]["total"].(float64); ok {
			count = int64(val)
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"user_id":        userID,
		"follower_count": count,
	})
}
func (ac *AuthController) GetAllFollowing(c echo.Context) error {
	// 1) ดึง user_id จาก param
	userStr := c.Param("id")
	userID, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user_id"})
	}

	// 2) Query จำนวน following
	rows, err := ac.AuthService.DBService.SelectData(
		"followers",                   // table
		[]string{"COUNT(*) AS total"}, // fields
		true,                          // where
		"follower_id = $1",            // where condition
		[]interface{}{userID},         // args
		false, "", "", "",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			echo.Map{"message": "Failed to get following count", "error": err.Error()})
	}

	// 3) อ่านผลลัพธ์
	var count int64 = 0
	if len(rows) > 0 {
		if val, ok := rows[0]["total"].(int64); ok {
			count = val
		} else if val, ok := rows[0]["total"].(float64); ok {
			count = int64(val)
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"user_id":         userID,
		"following_count": count,
	})
}

func (ac *AuthController) FollowUser(c echo.Context) error {
	// ดึง user_id ของตัวเองจาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}
	myuserID, _ := strconv.ParseInt(fmt.Sprintf("%v", uid), 10, 64)

	// ดึง user_id ของคนที่ฟอล จาก form-data
	userIDStr := c.FormValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid following user id"})
	}
	if userID < 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid format"})
	}

	// เพิ่ม follow
	data := map[string]interface{}{
		"follower_id":  myuserID,
		"following_id": userID,
	}
	if myuserID == userID {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "You cannot follow yourself"})
	}

	_, err = ac.AuthService.DBService.InsertData("followers", data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to follow user", "error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "follow successfully"})
}

func (ac *AuthController) UnFollowUser(c echo.Context) error {
	// 1)  ดึง user_id ของตัวเองจาก token
	uid := c.Get("user_id")
	if uid == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}
	myuserID, _ := strconv.ParseInt(fmt.Sprintf("%v", uid), 10, 64)

	// 2) ดึง following_id จาก param
	followingIDStr := c.Param("id")
	followingID, err := strconv.ParseInt(followingIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid following_id"})
	}
	if myuserID == followingID {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "You cannot unfollow yourself"})
	}
	rowsAffected, err := ac.AuthService.DBService.DeleteData(
		"followers",
		"follower_id = $1 AND following_id = $2",
		[]interface{}{myuserID, followingID},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to delete follower", "error": err.Error()})
	}
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "Follower not found"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Follow deleted successfully"})
}
