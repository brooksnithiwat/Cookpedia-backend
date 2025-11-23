package controllers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (ac *AuthController) GetAllBadges(c echo.Context) error {
	// 1) ดึง user_id จาก param
	userStr := c.Param("id")
	userID, err := strconv.ParseInt(userStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid user_id"})
	}

	badges := []map[string]interface{}{}

	// 2) ดึงข้อมูลจำนวนโพสต์
	postRows, err := ac.AuthService.DBService.SelectData(
		"posts",
		[]string{"COUNT(*) AS total_posts"},
		true,
		"user_id = $1",
		[]interface{}{userID},
		false, "", "", "",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get post count", "error": err.Error()})
	}
	totalPosts := int64(0)
	if len(postRows) > 0 {
		if val, ok := postRows[0]["total_posts"].(int64); ok {
			totalPosts = val
		} else if val, ok := postRows[0]["total_posts"].(float64); ok {
			totalPosts = int64(val)
		}
	}

	// 3) ดึงข้อมูล rating
	ratingRows, err := ac.AuthService.DBService.SelectData(
		"rating",
		[]string{"COUNT(*) AS five_star"},
		true,
		"user_id = $1 AND rate >= 5",
		[]interface{}{userID},
		false, "", "", "",
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to get rating", "error": err.Error()})
	}
	fiveStarCount := int64(0)
	if len(ratingRows) > 0 {
		if val, ok := ratingRows[0]["five_star"].(int64); ok {
			fiveStarCount = val
		} else if val, ok := ratingRows[0]["five_star"].(float64); ok {
			fiveStarCount = int64(val)
		}
	}

	// 4) ตรวจสอบ badge แบบ switch-case
	badgeIDs := []int{1, 2, 3, 4}
	for _, badgeID := range badgeIDs {
		switch badgeID {
		case 1: // first post
			if totalPosts >= 1 {
				badges = append(badges, map[string]interface{}{"badge_id": 1, "name": "first post"})
			}
		case 2: // got 5 star rate
			if fiveStarCount > 0 {
				badges = append(badges, map[string]interface{}{"badge_id": 2, "name": "got 5 star rate"})
			}
		case 3: // 10 recipe posted
			if totalPosts >= 10 {
				badges = append(badges, map[string]interface{}{"badge_id": 3, "name": "10 recipe posted"})
			}
		case 4: // 25 recipe posted
			if totalPosts >= 25 {
				badges = append(badges, map[string]interface{}{"badge_id": 4, "name": "25 recipe posted"})
			}
		}
	}

	// 5) Return JSON
	return c.JSON(http.StatusOK, echo.Map{
		"user_id": userID,
		"badges":  badges,
	})
}
