package models

type UserProfile struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	AboutMe   string `json:"aboutme"`
	ImageURL  string `json:"image_url"`
}
