package framework

import (
	"log"
	"net/http"

	"github.com/xchwan/simple-web-framework/framework/scope"
)

// Router 持有一組 HttpHandler，對每個進來的請求依序嘗試。
// 當某個 handler 回傳 Handled 後即停止，否則依匹配結果呼叫 errorHandler。
// Router 實作 http.Handler，可直接傳入 http.ListenAndServe。
type Router struct {
	handlers     []HttpHandler
	errorHandler ErrorHandlerFunc
	container    *Container
}

// NewRouter 建立並回傳一個空的 Router，預設使用標準錯誤處理。
func NewRouter() *Router {
	return &Router{
		errorHandler: defaultErrorHandler,
	}
}

// SetErrorHandler 設定自訂錯誤處理，覆蓋預設行為。
func (ro *Router) SetErrorHandler(f ErrorHandlerFunc) {
	ro.errorHandler = f
}

// UseContainer 設定此 Router 使用的 IoC Container。
func (ro *Router) UseContainer(c *Container) {
	ro.container = c
}

// Register 將一個 handler 加入路由表。
func (ro *Router) Register(h HttpHandler) {
	ro.handlers = append(ro.handlers, h)
}

// GET 註冊一個 GET 路由。
func (ro *Router) GET(path string, f HandlerFunc) {
	ro.Register(NewPathHandler(path, NewMethodHandler(http.MethodGet, f)))
}

// POST 註冊一個 POST 路由。
func (ro *Router) POST(path string, f HandlerFunc) {
	ro.Register(NewPathHandler(path, NewMethodHandler(http.MethodPost, f)))
}

// PUT 註冊一個 PUT 路由。
func (ro *Router) PUT(path string, f HandlerFunc) {
	ro.Register(NewPathHandler(path, NewMethodHandler(http.MethodPut, f)))
}

// DELETE 註冊一個 DELETE 路由。
func (ro *Router) DELETE(path string, f HandlerFunc) {
	ro.Register(NewPathHandler(path, NewMethodHandler(http.MethodDelete, f)))
}

// PATCH 註冊一個 PATCH 路由。
func (ro *Router) PATCH(path string, f HandlerFunc) {
	ro.Register(NewPathHandler(path, NewMethodHandler(http.MethodPatch, f)))
}

// Run 啟動 HTTP server 並監聽指定 addr（如 ":8080"）。
func (ro *Router) Run(addr string) error {
	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, ro)
}

// ServeHTTP 實作 http.Handler 介面，由 Go 的 HTTP server 在每個請求進來時自動呼叫。
func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = storeErrorHandler(r, ro.errorHandler)
	r = scope.InjectRequestScopeStore(r)
	if ro.container != nil {
		r = injectContainer(r, ro.container)
	}
	best := NotMatched
	for _, h := range ro.handlers {
		result := h.Handle(w, r)
		if result == Handled {
			return
		}
		if result > best {
			best = result
		}
	}
	switch best {
	case PathMatched:
		ro.errorHandler(w, r, http.StatusMethodNotAllowed)
	default:
		ro.errorHandler(w, r, http.StatusNotFound)
	}
}
