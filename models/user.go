package models

import "time"

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Password       string    `json:"password,omitempty"` // omitempty for OAuth users
	GoogleID       string    `json:"google_id,omitempty"`
	ProfilePicture string    `json:"profile_picture,omitempty"`
	Provider       string    `json:"provider"` // "local", "google"
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}
