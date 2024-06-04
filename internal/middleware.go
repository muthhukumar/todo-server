package internal

import (
	"context"
	"net/http"
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
