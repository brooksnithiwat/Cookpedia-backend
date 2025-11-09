package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "github.com/lib/pq"
)


func (s *DatabaseService) ParseDateTime(input interface{}, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Println("❌ Failed to load timezone, defaulting to UTC:", err)
		loc = time.UTC
	}
	var t time.Time
	if input != nil {
		switch v := input.(type) {
		case time.Time:
			t = v.In(loc)
		case string:
			parsed, err := time.Parse(time.RFC3339, v)
			if err != nil {
				log.Println("❌ Failed to parse datetime string:", err)
				t = time.Now().In(loc)
			} else {
				t = parsed.In(loc)
			}
		default:
			log.Println("❌ Unexpected type for datetime:", reflect.TypeOf(v))
			t = time.Now().In(loc)
		}
	} else {
		log.Println("⚠️ datetime is nil, using now() instead")
		t = time.Now().In(loc)
	}

	return t
}

func (s *DatabaseService) GetPostWithTagsAndDetails(postID int) (map[string]interface{}, error) {
	post := make(map[string]interface{})

	// ✅ 1) ดึงข้อมูล post (เพิ่ม created_at)
	row := s.DB.QueryRow(`
		SELECT post_id, user_id, menu_name, story, image_url, created_at
		FROM posts
		WHERE post_id = $1
	`, postID)

	var userID int
	var menuName, story, imageURL sql.NullString
	var createdAt time.Time

	err := row.Scan(&postID, &userID, &menuName, &story, &imageURL, &createdAt)
	if err != nil {
		return nil, err
	}

	// ✅ ใส่ข้อมูลลง map (อันที่ขาดมาก่อน: user_id + created_at)
	post["post_id"] = postID
	post["user_id"] = userID
	post["menu_name"] = menuName.String
	post["story"] = story.String
	post["image_url"] = imageURL.String
	post["created_at"] = createdAt

	// ✅ 2) ดึง categories_tags
	cats := []string{}
	rows, _ := s.DB.Query(`
		SELECT ct.category_tag_name
		FROM post_categories pc
		JOIN categories_tag ct ON pc.category_tag_id = ct.category_tag_id
		WHERE pc.post_id = $1
	`, postID)
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		cats = append(cats, name)
	}
	post["categories_tags"] = cats

	// ✅ 3) ดึง ingredients_tags
	ingTags := []string{}
	rows2, _ := s.DB.Query(`
		SELECT it.ingredient_tag_name
		FROM post_ingredients pi
		JOIN ingredients_tag it ON pi.ingredient_tag_id = it.ingredient_tag_id
		WHERE pi.post_id = $1
	`, postID)
	defer rows2.Close()

	for rows2.Next() {
		var name string
		rows2.Scan(&name)
		ingTags = append(ingTags, name)
	}
	post["ingredients_tags"] = ingTags

	// ✅ 4) ดึง ingredients_detail
	ings := []string{}
	rows3, _ := s.DB.Query(`
		SELECT detail
		FROM ingredients_detail
		WHERE post_id = $1
	`, postID)
	defer rows3.Close()

	for rows3.Next() {
		var detail string
		rows3.Scan(&detail)
		ings = append(ings, detail)
	}
	post["ingredients"] = ings

	// ✅ 5) ดึง instructions
	ins := []string{}
	rows4, _ := s.DB.Query(`
		SELECT detail
		FROM instructions
		WHERE post_id = $1
		ORDER BY step_number ASC
	`, postID)
	defer rows4.Close()

	for rows4.Next() {
		var step string
		rows4.Scan(&step)
		ins = append(ins, step)
	}
	post["instructions"] = ins

	return post, nil
}

// DatabaseService.go
func (db *DatabaseService) GetPostIDsByUser(userID int64) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT post_id FROM posts WHERE user_id = $1 ORDER BY post_id DESC"
	rows, err := db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		postIDs = append(postIDs, id)
	}

	return postIDs, nil
}
func (s *DatabaseService) GetAllPostIDs() ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT post_id FROM posts ORDER BY post_id DESC"
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		fmt.Println("[DEBUG] GetAllPostIDs query error:", err)
		return nil, err
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			fmt.Println("[DEBUG] GetAllPostIDs scan error:", err)
			continue
		}
		postIDs = append(postIDs, id)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("[DEBUG] GetAllPostIDs rows error:", err)
		return nil, err
	}

	return postIDs, nil
}
