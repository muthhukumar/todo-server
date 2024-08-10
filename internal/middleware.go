package internal

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"todo-server/models"
	"todo-server/utils"

	"github.com/go-chi/chi/v5"
)

func ExtractTaskId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")

		id, err := strconv.Atoi(idStr)

		if err != nil {
			utils.JsonResponse(w, http.StatusBadRequest, models.MsgResponse{Message: "Invalid Task ID"})
			return
		}
		ctx := context.WithValue(r.Context(), "taskId", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthWithApiKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		xAPIKey := r.Header.Get("x-api-key")

		configuredApiKey := os.Getenv("API_KEY")

		utils.Assert(configuredApiKey != "", "API_KEY should not be empty")

		if xAPIKey == "" || xAPIKey != configuredApiKey {
			utils.JsonResponse(w, http.StatusUnauthorized, models.MsgResponse{Message: "Invalid API key"})
			return
		}

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}
