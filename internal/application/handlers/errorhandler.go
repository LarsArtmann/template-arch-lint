package handlers

import (
	"encoding/json"
	"net/http"
)

func sendErrorResponse(w http.ResponseWriter, httpStatus int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
