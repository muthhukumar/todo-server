package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"todo-server/api"

	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")

	if connStr == "" {
		log.Fatal("POSTGRES_CONNECTION_STRING environment variable not set")
	}

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		fmt.Println("Failed to connect to database", err)
		return
	}
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	api.SetupRoutes(r, db)

	err = db.Ping()
	if err != nil {
		fmt.Println("Ping failed", err)
		return
	}

	fmt.Println("Connected to the database")

	http.ListenAndServe(":3000", r)
}
