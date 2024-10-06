package internal

import (
	"database/sql"
	"log"
	"os"
	"todo-server/utils"

	_ "github.com/lib/pq"
)

func SetupDatabase() *sql.DB {
	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")

	utils.Assert(connStr != "", "POSTGRES_CONNECTION_STRING environment variable is set")

	db, err := sql.Open("postgres", connStr)

	// db.SetMaxOpenConns(5)
	// db.SetMaxIdleConns(5)
	// db.SetConnMaxLifetime(5 * time.Minute)
	// db.SetMaxOpenConns(25)
	// db.SetMaxIdleConns(25)
	// db.SetConnMaxLifetime(5 * time.Minute)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ping to database failed: %v", err)
	}

	log.Println("Database connection established successfully")

	return db
}
