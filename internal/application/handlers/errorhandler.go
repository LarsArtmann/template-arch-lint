package handlers

import (
	"encoding/json/v2"
	"net/http"
)

func sendErrorResponse(w http.ResponseWriter, httpStatus int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.MarshalWrite(w, map[string]string{"error": message})
}
