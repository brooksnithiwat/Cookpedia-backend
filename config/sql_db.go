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

	// ❌ ลบ statement_cache_mode เพราะ lib/pq ไม่รองรับ
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		dbHost, dbUser, dbPassword, dbName, dbPort,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB (sql):", err)
	}

	// ตั้งค่า connection pool
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(1 * time.Hour)

	// Ping DB
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping DB (sql):", err)
	}

	SQLDB = db
	log.Println("Connected to PostgreSQL (database/sql) with connection pool")
}
