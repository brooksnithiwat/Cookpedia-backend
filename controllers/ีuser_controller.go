package controllers

import (
	"fmt"
	"go-auth/models"
	"go-auth/supabaseutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (ac *AuthController) EditUserProfile(c echo.Context) error {
	userID := c.Get("user_id") //get ค่า user_id จากใน token

	if userID == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User did not authenticated"})
	}

	firstname := c.FormValue("firstname")
	lastname := c.FormValue("lastname")
	phone := c.FormValue("phone")
	email := c.FormValue("email")
	aboutme := c.FormValue("aboutme")
	file, err := c.FormFile("image")
	var imageURL string
	if file != nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to open image"})
		}
		defer src.Close()
		imageURL, err = supabaseutil.UploadFile(src, file, userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to upload image", "error": err.Error()})
		}
		if err := ac.AuthService.UpdateUserImage(userID, imageURL); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to update user profile"})
		}
	}

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
	whereCon := fmt.Sprintf("user_id = $%d", len(fields)+1)
	values = append(values, userID)
	_, err = ac.AuthService.DBService.UpdateData("users", fields, whereCon, values)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to update user profile"})
	}
	resp := map[string]interface{}{}
	for i, f := range fields {
		resp[f] = values[i]
	}
	return c.JSON(http.StatusOK,"Edit Profile Success")
}

func (ac *AuthController) GetUserProfile(c echo.Context) error {
	userID := c.Get("user_id")
	if userID == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "User not authenticated"})
	}
	fields := []string{"firstname", "lastname", "email", "phone", "aboutme", "image_url"}
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
		Firstname: row["firstname"].(string),
		Lastname:  row["lastname"].(string),
		Email:     row["email"].(string),
		Phone:     row["phone"].(string),
		AboutMe:   row["aboutme"].(string),
		ImageURL:  row["image_url"].(string),
	}
	return c.JSON(http.StatusOK, profile)
}
