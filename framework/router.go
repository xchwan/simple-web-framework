package framework

import (
	"context"
	"log"
	"net/http"

	"github.com/xchwan/simple-web-framework/framework/plugin"
	"github.com/xchwan/simple-web-framework/framework/routing"
	"github.com/xchwan/simple-web-framework/framework/scope"
)

// HandlerFunc 是 routing.HandlerFunc 的型別別名，讓使用者不需要直接 import routing 套件。
type HandlerFunc = routing.HandlerFunc

// PathParam 從 request context 取出指定的 path parameter。
func PathParam(r *http.Request, key string) string {
	return routing.PathParam(r, key)
}

// Router 持有一組 HttpHandler，對每個進來的請求依序嘗試。
// 當某個 handler 回傳 Handled 後即停止，否則依匹配結果呼叫 errorHandler。
// Router 實作 http.Handler，可直接傳入 http.ListenAndServe。
type Router struct {
	handlers     []routing.HttpHandler
	errorHandler ErrorHandlerFunc
	container    *Container
	codecs       map[string]plugin.Codec
}

// NewRouter 建立並回傳一個空的 Router，預設使用標準錯誤處理，並內建 JSON 與 text/plain 的 Codec。
func NewRouter() *Router {
	r := &Router{
		errorHandler: defaultErrorHandler,
		container:    NewContainer(),
		codecs:       make(map[string]plugin.Codec),
	}
	r.RegisterCodec("application/json", &jsonCodec{})
	r.RegisterCodec("text/plain", &textCodec{})
	return r
}

// AddPlugin 安裝一個插件，插件可藉此向 Router 註冊 Codec 等擴充。
func (ro *Router) AddPlugin(p plugin.Plugin) {
	p.Install(ro)
}

// RegisterCodec 向 Router 註冊指定 media type 的 Codec，實作 plugin.Registrar。
func (ro *Router) RegisterCodec(mediaType string, c plugin.Codec) {
	ro.codecs[mediaType] = c
}

// SetErrorHandler 設定自訂錯誤處理，覆蓋預設行為。
func (ro *Router) SetErrorHandler(f ErrorHandlerFunc) {
	ro.errorHandler = f
}

// Bind 向容器註冊一個依賴，s 省略時預設使用 SingletonScope。
// 新增自訂 scope 只需傳入對應的 scope.Scope 實作，不需修改 Router。
func (ro *Router) Bind(name string, factory func() any, s ...scope.Scope) {
	ro.container.Register(name, factory, s...)
}

// Resolve 從容器取得指定名稱的依賴實體，供啟動時組裝使用。
func (ro *Router) Resolve(name string) any {
	return ro.container.Resolve(context.Background(), name)
}

// register 將一個 HttpHandler 加入路由表。
func (ro *Router) register(h routing.HttpHandler) {
	ro.handlers = append(ro.handlers, h)
}

// GET 註冊一個 GET 路由。
func (ro *Router) GET(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodGet, f)))
}

// POST 註冊一個 POST 路由。
func (ro *Router) POST(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPost, f)))
}

// PUT 註冊一個 PUT 路由。
func (ro *Router) PUT(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPut, f)))
}

// DELETE 註冊一個 DELETE 路由。
func (ro *Router) DELETE(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodDelete, f)))
}

// PATCH 註冊一個 PATCH 路由。
func (ro *Router) PATCH(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPatch, f)))
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
	r = injectCodecs(r, ro.codecs)
	if ro.container != nil {
		r = injectContainer(r, ro.container)
	}
	best := routing.NotMatched
	for _, h := range ro.handlers {
		result := h.Handle(w, r)
		if result == routing.Handled {
			return
		}
		if result > best {
			best = result
		}
	}
	switch best {
	case routing.PathMatched:
		ro.errorHandler(w, r, http.StatusMethodNotAllowed)
	default:
		ro.errorHandler(w, r, http.StatusNotFound)
	}
}
