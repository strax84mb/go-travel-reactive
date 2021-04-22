package web

import (
	"encoding/json"
	"net/http"
)

func Ok(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func BadRequest(w http.ResponseWriter, message string, details map[string][]string) {
	writeError(w, http.StatusBadRequest, message, details)
}

func InternalServerError(w http.ResponseWriter, message string, details map[string][]string) {
	writeError(w, http.StatusInternalServerError, message, details)
}

func writeError(w http.ResponseWriter, status int, message string, details map[string][]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Message: message,
		Details: details,
	})
}

type errorResponse struct {
	Message string              `json:"message"`
	Details map[string][]string `json:"details,omitempty"`
}
