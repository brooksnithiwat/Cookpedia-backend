package models

type UserProfile struct {
	UserID    int64  `json:"user_id" form:"user_id"`
	UserName  string `json:"username" form:"username"`
	Firstname string `json:"firstname" form:"firstname"`
	Lastname  string `json:"lastname" form:"lastname"`
	Email     string `json:"email" form:"email"`
	Phone     string `json:"phone" form:"phone"`
	AboutMe   string `json:"aboutme" form:"aboutme"`
	ImageURL  string `json:"image_url" form:"image_url"`
}
