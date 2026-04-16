package framework

import (
	"context"
	"encoding/json"
	"net/http"
)

// ErrorHandlerFunc 統一處理所有 HTTP 錯誤回應。
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, statusCode int)

// errorHandlerKey 是在 context 中存取 ErrorHandlerFunc 的 key。
type errorHandlerKey struct{}

// storeErrorHandler 將 ErrorHandlerFunc 存入 request context。
func storeErrorHandler(r *http.Request, f ErrorHandlerFunc) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), errorHandlerKey{}, f))
}

// loadErrorHandler 從 request context 取出 ErrorHandlerFunc。
func loadErrorHandler(r *http.Request) ErrorHandlerFunc {
	if f, ok := r.Context().Value(errorHandlerKey{}).(ErrorHandlerFunc); ok {
		return f
	}
	return defaultErrorHandler
}

// ErrorBody 是統一的 JSON 錯誤回應格式。
type ErrorBody struct {
	Message string `json:"message"`
}

// defaultErrorHandler 是預設的錯誤處理，回傳 JSON 格式的錯誤訊息。
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorBody{Message: http.StatusText(statusCode)})
}
