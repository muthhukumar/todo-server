package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func SetupDatabase() *sql.DB {
	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")

	if connStr == "" {
		log.Fatal("POSTGRES_CONNECTION_STRING environment variable not set")
	}

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ping to database failed: %v", err)
	}

	fmt.Println("Database connection established successfully")

	return db
}
