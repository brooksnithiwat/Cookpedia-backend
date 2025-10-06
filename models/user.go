// package models

// type User struct {
// 	ID        int     `gorm:"column:user_id;primaryKey" json:"user_id"`
// 	Username  string  `json:"username"`
// 	Email     string  `json:"email"`
// 	Password  string  `json:"password"`
// 	GoogleID  *string `json:"google_id"`
// 	Provider  string  `json:"provider"`
// 	Role      string  `json:"role"`
// 	Firstname string  `json:"firstname"`
// 	Lastname  string  `json:"lastname"`
// 	Phone     string  `json:"phone"`
// 	Aboutme   string  `json:"aboutme"`
// 	ImageURL  string  `json:"image_url"`
// }

//	type GoogleUserInfo struct {
//		ID    string `json:"id"`
//		Email string `json:"email"`
//		Name  string `json:"name"`
//	}
package models

type User struct {
	ID        int     `gorm:"column:user_id;primaryKey" json:"user_id" form:"user_id"`
	Username  string  `json:"username" form:"username"`
	Email     string  `json:"email" form:"email"`
	Password  string  `json:"password" form:"password"`
	GoogleID  *string `json:"google_id" form:"google_id"`
	Provider  string  `json:"provider" form:"provider"`
	Role      string  `json:"role" form:"role"`
	Firstname string  `json:"firstname" form:"firstname"`
	Lastname  string  `json:"lastname" form:"lastname"`
	Phone     string  `json:"phone" form:"phone"`
	Aboutme   string  `json:"aboutme" form:"aboutme"`
	ImageURL  string  `json:"image_url" form:"image_url"`
}

//	type GoogleUserInfo struct {
//		ID    string `json:"user_id"`
//		Email string `json:"email"`
//		Name  string `json:"name"`
//	}
type GoogleUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	ImageURL string `json:"picture"` // ต้องตรงกับชื่อ field จาก Google API (ดึง url รูปโปรไฟล์ด้วย)
}
