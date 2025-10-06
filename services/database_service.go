package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-auth/models"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseService ให้บริการติดต่อฐานข้อมูลด้วย SQL ตรง
// ต้องแนบ *sql.DB ตอนสร้าง

type DatabaseService struct {
	DB *sql.DB
}

func NewDatabaseService(db *sql.DB) *DatabaseService {
	return &DatabaseService{DB: db}
}

func (s *DatabaseService) GetUserByUsernameOrEmail(usernameOrEmail string) (*models.User, error) {
	var user models.User
	query := `
		SELECT user_id, username, email, password, provider, role
		FROM users
		WHERE username = $1 OR email = $1
		LIMIT 1
	`
	row := s.DB.QueryRow(query, usernameOrEmail)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Provider, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Register สมัครสมาชิกใหม่
func (s *DatabaseService) Register(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists int
	err := s.DB.QueryRowContext(ctx, `SELECT COUNT(1) FROM users WHERE username = $1 OR email = $2`, user.Username, user.Email).Scan(&exists)
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("user already exists")
	}

	query := `INSERT INTO users (username, email, password, provider, role, firstname, lastname, phone, aboutme, image_url) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err = s.DB.ExecContext(ctx, query,
		user.Username, user.Email, user.Password, user.Provider, user.Role, user.Firstname, user.Lastname, user.Phone, user.Aboutme, user.ImageURL,
	)
	return err
}

// Login ตรวจสอบ username/email และ password
func (s *DatabaseService) Login(usernameOrEmail, password string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT user_id, username, email, password, google_id, provider, role, firstname, lastname, phone, aboutme, image_url FROM users WHERE (username = $1 OR email = $1) AND provider = 'local'`
	row := s.DB.QueryRowContext(ctx, query, usernameOrEmail)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.GoogleID, &user.Provider, &user.Role, &user.Firstname, &user.Lastname, &user.Phone, &user.Aboutme, &user.ImageURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}
	// ตรวจสอบ password (hashing logic ควรอยู่ที่ auth_service)
	return &user, nil
}

func (s *DatabaseService) SelectData(
	table string,
	fields []string,
	where bool,
	whereCon string,
	whereArgs []interface{},
	join bool,
	joinTable string,
	joinCondition string,
	orderAndLimit string,
) ([]map[string]interface{}, error) {
	// สร้าง query string
	query := "SELECT "
	if len(fields) == 0 {
		query += "*"
	} else {
		query += strings.Join(fields, ", ")
	}
	query += " FROM " + table

	if join && joinTable != "" {
		query += " JOIN " + joinTable
		if joinCondition != "" {
			query += " ON " + joinCondition
		}
	}

	if where && whereCon != "" {
		// เปลี่ยน ? เป็น $1, $2, ... สำหรับ PostgreSQL
		q := whereCon
		for i := 1; strings.Contains(q, "?"); i++ {
			q = strings.Replace(q, "?", fmt.Sprintf("$%d", i), 1)
		}
		query += " WHERE " + q
	}

	if orderAndLimit != "" {
		query += " " + strings.TrimSpace(orderAndLimit)
	}
	query = strings.TrimSpace(query)
	// Remove trailing WHERE, AND, OR if whereCon is empty
	query = strings.TrimSuffix(query, " WHERE")
	query = strings.TrimSuffix(query, " AND")
	query = strings.TrimSuffix(query, " OR")
	fmt.Println("Executing query:", query)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.DB.QueryContext(ctx, query, whereArgs...)
	if err != nil {
		fmt.Println("[DEBUG] DB Query error:", err)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Println("[DEBUG] DB Columns error:", err)
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// handle []byte to string
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("[DEBUG] DB Rows error:", err)
		return nil, err
	}
	return results, nil
}

// UpdateData อัปเดตข้อมูลใน table ตาม fields ที่กำหนด (field values ตามลำดับใน fields) และ where condition
func (s *DatabaseService) UpdateData(
	table string,
	fields []string,
	condition string,
	values []interface{}, // field values ตามลำดับใน fields ต่อด้วย conditionValues
) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var setClauses []string
	for i, field := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, i+1))
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setClauses, ", "), condition)
	fmt.Println("Executing query:", query)
	fmt.Println("With values:", values)

	result, err := s.DB.ExecContext(ctx, query, values...)
	if err != nil {
		fmt.Println("[DEBUG] UpdateData error:", err)
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("[DEBUG] RowsAffected error:", err)
		return 0, err
	}

	return rowsAffected, nil
}

// InsertData แทรกข้อมูลใหม่เข้า table ที่กำหนด
func (s *DatabaseService) InsertData(table string, data map[string]interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var columns []string
	var placeholders []string
	var values []interface{}

	i := 1
	for column, value := range data {
		columns = append(columns, column)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, value)
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	fmt.Println("Executing query:", query)
	fmt.Println("With values:", values)

	result, err := s.DB.ExecContext(ctx, query, values...)
	if err != nil {
		fmt.Println("[DEBUG] InsertData error:", err)
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("[DEBUG] RowsAffected error:", err)
		return 0, err
	}

	return rowsAffected, nil
}

// DeleteData ลบข้อมูลจาก table ตามเงื่อนไขที่กำหนด
func (s *DatabaseService) DeleteData(table string, condition string, conditionValues []interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, condition)
	fmt.Println("Executing query:", query)
	fmt.Println("With values:", conditionValues)

	result, err := s.DB.ExecContext(ctx, query, conditionValues...)
	if err != nil {
		fmt.Println("[DEBUG] DeleteData error:", err)
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("[DEBUG] RowsAffected error:", err)
		return 0, err
	}

	return rowsAffected, nil
}

func (s *DatabaseService) SelectJoin(
	baseTable string,
	fields []string, // fields ที่จะ select เช่น ["p.post_id","p.menu_name","c.category_tag_name"]
	joins []string, // join clauses เช่น ["JOIN post_categories pc ON p.post_id=pc.post_id","JOIN categories_tag c ON pc.category_tag_id=c.category_tag_id"]
	condition string, // WHERE condition เช่น "p.user_id=$1"
	conditionValues []interface{}, // values สำหรับ WHERE condition
	orderBy string, // ORDER BY clause (optional)
	limit int, // LIMIT (0=ไม่จำกัด)
	offset int, // OFFSET (0=ไม่จำกัด)
) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	selectClause := "*"
	if len(fields) > 0 {
		selectClause = strings.Join(fields, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", selectClause, baseTable)

	if len(joins) > 0 {
		query += " " + strings.Join(joins, " ")
	}

	if condition != "" {
		query += " WHERE " + condition
	}

	if orderBy != "" {
		query += " ORDER BY " + orderBy
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	fmt.Println("Executing query:", query)
	fmt.Println("With values:", conditionValues)

	rows, err := s.DB.QueryContext(ctx, query, conditionValues...)
	if err != nil {
		fmt.Println("[DEBUG] SelectJoin error:", err)
		return nil, err
	}

	return rows, nil
}

func (s *DatabaseService) GetPostWithTagsAndDetails(postID int) (map[string]interface{}, error) {
	post := make(map[string]interface{})

	// 1) ดึง post
	row := s.DB.QueryRow("SELECT post_id, user_id, menu_name, story, image_url FROM posts WHERE post_id = $1", postID)
	var userID int
	var menuName, story, imageURL sql.NullString
	err := row.Scan(&postID, &userID, &menuName, &story, &imageURL)
	if err != nil {
		return nil, err
	}

	post["post_id"] = postID
	post["menu_name"] = menuName.String
	post["story"] = story.String
	post["image_url"] = imageURL.String

	// 2) ดึง categories_tags
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

	// 3) ดึง ingredients_tags
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

	// 4) ดึง ingredients_detail
	ings := []string{}
	rows3, _ := s.DB.Query("SELECT detail FROM ingredients_detail WHERE post_id = $1", postID)
	defer rows3.Close()
	for rows3.Next() {
		var detail string
		rows3.Scan(&detail)
		ings = append(ings, detail)
	}
	post["ingredients"] = ings

	// 5) ดึง instructions
	ins := []string{}
	rows4, _ := s.DB.Query("SELECT detail FROM instructions WHERE post_id = $1 ORDER BY step_number ASC", postID)
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
