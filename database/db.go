package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := "postgres://user:password@localhost:5432/car_listing?sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Successfully connected to the database")
	runMigrations()
}

func runMigrations() {
	migrationFile := "migrations/001_create_cars_table.sql"
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	_, err = DB.Exec(string(content))
	if err != nil {
		log.Fatal("Failed to run migration:", err)
	}

	fmt.Println("Database migration executed successfully")
}
