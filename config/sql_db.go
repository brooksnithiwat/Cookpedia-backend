package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB (sql):", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping DB (sql):", err)
	}
	SQLDB = db
	log.Println("Connected to PostgreSQL (database/sql)")
}
