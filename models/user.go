package models

type User struct {
	ID        int     `gorm:"column:user_id;primaryKey" json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	GoogleID  *string `json:"google_id"`
	Provider  string  `json:"provider"`
	Role      string  `json:"role"`
	Firstname string  `json:"firstname"`
	Lastname  string  `json:"lastname"`
	Phone     string  `json:"phone"`
	Aboutme   string  `json:"aboutme"`
	ImageURL  string  `json:"image_url"`
}

type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
