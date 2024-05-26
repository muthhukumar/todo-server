package utils

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, code int, response any) {
	w.Header().Set("Content-Type", "applications/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(response)
}
