package framework

import "net/http"

// MethodHandler 只在請求的 HTTP Method 符合時才將請求往下傳遞。
type MethodHandler struct {
	method  string
	wrapped HttpHandler
}

// NewMethodHandler 建立一個 MethodHandler，包裝 wrapped handler。
func NewMethodHandler(method string, wrapped HttpHandler) *MethodHandler {
	return &MethodHandler{method: method, wrapped: wrapped}
}

// Handle 實作 HttpHandler。Method 不符回傳 PathMatched（路徑已符合但方法不對），符合則交給 wrapped。
func (d *MethodHandler) Handle(w http.ResponseWriter, r *http.Request) HandleResult {
	if r.Method != d.method {
		return PathMatched
	}
	return d.wrapped.Handle(w, r)
}
