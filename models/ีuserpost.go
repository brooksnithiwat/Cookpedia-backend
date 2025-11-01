package models

type OwnerPost struct {
	ProfileImage string `json:"profile_image"`
	Username     string `json:"username"`
	CreatedDate string  `json:"created_date"` 
	CreatedTime string  `json:"created_time"` 
}

type PostDetail struct {
	PostID          int      `json:"post_id"`
	MenuName        string   `json:"menu_name"`
	Story           string   `json:"story"`
	ImageURL        string   `json:"image_url"`
	CategoriesTags  []string `json:"categories_tags"`
	IngredientsTags []string `json:"ingredients_tags"`
	Ingredients     []string `json:"ingredients"`
	Instructions    []string `json:"instructions"`
}

type PostResponse struct {
	OwnerPost OwnerPost  `json:"owner_post"`
	Post      PostDetail `json:"post"`
}
