package models

import "time"

// Comment struct สำหรับเก็บข้อมูล comment
type Comment struct {
	CommentID int64     `json:"comment_id"`
	UserID    int64     `json:"user_id"`
	PostID    int64     `json:"post_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CommentResponse struct สำหรับ response ที่รวมข้อมูล user
type CommentResponse struct {
	UserID     int64  `json:"user_id"`
	Username   string `json:"username"`
	ProfileImg string `json:"profile_img"`
	CreatedAt  string `json:"created_at"` // format เป็น string เช่น "2006-01-02 15:04:05"
	CommentID  int64  `json:"comment_id"`
	Content    string `json:"content"`
}
