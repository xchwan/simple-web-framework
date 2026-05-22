package builtin

import (
	"encoding/json"
	"net/http"
)

// DefaultErrorHandler responds with a JSON error body for routing-layer errors (404, 405).
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{Message: http.StatusText(statusCode)})
}
