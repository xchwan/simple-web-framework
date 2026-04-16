package framework

import (
	"encoding/json"
	"net/http"
)

// Respond 是統一的回應入口，將 body 以 JSON 編碼回傳。
// 204 不設 Content-Type、不帶 body。
// routing 層的 404/405 由 Router.ServeHTTP 直接呼叫 errorHandler，不走這裡。
func Respond(w http.ResponseWriter, r *http.Request, statusCode int, body any) {
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
