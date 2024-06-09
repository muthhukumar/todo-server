package main

import (
	"fmt"
	"net/http"
	"os"
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

	var port string

	if port = os.Getenv("PORT"); port == "" {
		port = "3000"
	}

	db := internal.SetupDatabase()
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:1420", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "x-api-key"},
		AllowCredentials: true,
	})

	r.Use(c.Handler)

	api.SetupRoutes(r, db)

	emailAuth := internal.LoadEmailCredentials()

	internal.SetupCronJobs(db, emailAuth)

	addr := fmt.Sprintf(":%v", port)

	http.ListenAndServe(addr, r)
}
