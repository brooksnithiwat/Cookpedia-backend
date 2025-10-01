package models

type PostResponse struct {
	PostID          int      `json:"post_id"`
	MenuName        string   `json:"menu_name"`
	Story           string   `json:"story"`
	ImageURL        string   `json:"image_url"`
	CategoriesTags  []string `json:"categories_tags"`
	IngredientsTags []string `json:"ingredients_tags"`
	Ingredients     []string `json:"ingredients"`  // ingredients_detail
	Instructions    []string `json:"instructions"` // instructions
}
