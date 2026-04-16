package framework

import (
	"encoding/json"
	"net/http"
)

// Respond 是統一的回應入口。
// statusCode < 400：直接回傳 JSON，body 為 nil 時只寫狀態碼。
// statusCode >= 400：交給 errorHandler 處理，body 忽略。
func Respond(w http.ResponseWriter, r *http.Request, statusCode int, body any) {
	if statusCode >= 400 {
		loadErrorHandler(r)(w, r, statusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if body != nil {
		json.NewEncoder(w).Encode(body)
	}
}
