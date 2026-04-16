package framework

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Respond 是統一的回應入口。
// statusCode < 400：回傳 JSON（204 不設 Content-Type、不帶 body）。
// statusCode >= 400：交給 errorHandler 處理，body 忽略。
func Respond(w http.ResponseWriter, r *http.Request, statusCode int, body any) {
	if statusCode >= 400 {
		loadErrorHandler(r)(w, r, statusCode)
		return
	}
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if body != nil {
		json.NewEncoder(w).Encode(body)
	}
}

// RespondText 回傳純文字回應，不走 errorHandler。
func RespondText(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	fmt.Fprint(w, message)
}
