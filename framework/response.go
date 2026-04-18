package framework

import (
	"net/http"

	"github.com/xchwan/simple-web-framework/framework/plugin"
)

// Respond 依 Content-Type header 選擇對應的 Codec，將 body 序列化後回傳。
// 204 不設 Content-Type、不帶 body。
// routing 層的 404/405 由 Router.ServeHTTP 直接呼叫 errorHandler，不走這裡。
func Respond(w http.ResponseWriter, r *http.Request, statusCode int, body any) {
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return
	}
	mt, c := plugin.Lookup(r, r.Header.Get("Content-Type"))
	w.Header().Set("Content-Type", mt)
	w.WriteHeader(statusCode)
	if body != nil {
		c.Encode(w, body)
	}
}
