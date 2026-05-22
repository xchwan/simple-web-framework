package builtin

import (
	"encoding/json"
	"net/http"
)

// DefaultErrorHandler 是預設的錯誤處理，回傳 JSON 格式的錯誤訊息。
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{Message: http.StatusText(statusCode)})
}
