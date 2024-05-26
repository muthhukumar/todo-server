package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"todo-server/api"
	"todo-server/internal"
)

func main() {
	internal.LoadDotEnvFile()

	db := internal.SetupDatabase()
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	api.SetupRoutes(r, db)

	internal.SetupCronJobs(db)

	// Enable it when email credentials is needed
	// internal.LoadEmailCredentials()

	http.ListenAndServe(":3000", r)
}
