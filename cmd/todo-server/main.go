package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/cors"

	"todo-server/api"
	"todo-server/internal"
)

func main() {
	internal.LoadDotEnvFile()

	db := internal.SetupDatabase()
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:1420", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	r.Use(c.Handler)

	api.SetupRoutes(r, db)

	// emailAuth := internal.LoadEmailCredentials()

	// internal.SetupCronJobs(db, emailAuth)

	// Enable it when email credentials is needed

	http.ListenAndServe(":3000", r)
}
