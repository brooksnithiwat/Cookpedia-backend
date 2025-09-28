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
