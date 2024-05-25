package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"todo-server/api"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	api.SetupRoutes(r)

	http.ListenAndServe(":3000", r)
}
