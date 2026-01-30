package database

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB holds the database connection
var DB *sql.DB

// InitDB establishes connection to PostgreSQL database
func InitDB(connectionString string) (*sql.DB, error) {
	log.Println("Connecting to database...")

	// Disable prepared statement cache for PgBouncer compatibility (Supabase)
	if strings.Contains(connectionString, "?") {
		connectionString += "&default_query_exec_mode=exec"
	} else {
		connectionString += "?default_query_exec_mode=exec"
	}

	// Open database with pgx driver
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	DB = db
	log.Println("Database connected successfully")
	return db, nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed")
	}
}
