package utils

import (
	"encoding/json"
	"net/http"
	"todo-server/models"
)

func RespondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "applications/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(models.MsgResponse{Message: message})
}
