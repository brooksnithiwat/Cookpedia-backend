package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var SQLDB *sql.DB

func InitSQLDB() {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		log.Fatal("Database environment variables are not set")
	}

	// ปิด prepared statement cache
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable statement_cache_mode=none",
		dbHost, dbUser, dbPassword, dbName, dbPort,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB (sql):", err)
	}

	// ตรวจสอบว่า DB ใช้งานได้
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping DB (sql):", err)
	}

	// ตั้งค่า connection pool
	db.SetMaxIdleConns(10)               // จำนวน connection idle สูงสุด
	db.SetMaxOpenConns(100)              // จำนวน connection เปิดสูงสุด
	db.SetConnMaxLifetime(1 * time.Hour) // connection แต่ละตัวมีอายุไม่เกิน 1 ชั่วโมง

	SQLDB = db
	log.Println("Connected to PostgreSQL (database/sql) with connection pool")
}
