package framework

import (
	"context"
	"log"
	"net/http"
	"reflect"

	"github.com/xchwan/simple-web-framework/framework/builtin"
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
type Router struct {
	handlers     []routing.HttpHandler
	errorHandler ErrorHandlerFunc
	container    *Container
	plugins      map[reflect.Type]any
}

// NewRouter 建立並回傳一個空的 Router，預設使用標準錯誤處理，並內建 JSON 與 text/plain 的 Codec。
func NewRouter() *Router {
	r := &Router{
		errorHandler: builtin.DefaultErrorHandler,
		container:    NewContainer(),
		plugins:      make(map[reflect.Type]any),
	}
	cr := NewCodecRegistry()
	cr.Register("application/json", &builtin.JsonCodec{})
	cr.Register("text/plain", &builtin.TextCodec{})
	r.plugins[reflect.TypeOf(cr)] = cr
	return r
}

// SetErrorHandler 設定自訂錯誤處理，覆蓋預設行為。
func (ro *Router) SetErrorHandler(f ErrorHandlerFunc) {
	ro.errorHandler = f
}

// AddPlugin 安裝一個插件，以型別為 key 存入 plugins map，同型別只會保留最後一個。
func (ro *Router) AddPlugin(p plugin.Plugin) {
	ro.plugins[reflect.TypeOf(p)] = p
	p.Install(ro)
}

// RegisterCodec 向內建 CodecRegistry 註冊指定 media type 的 Codec，實作 plugin.Registrar。
func (ro *Router) RegisterCodec(mediaType string, c plugin.Codec) {
	ro.plugins[reflect.TypeOf((*CodecRegistry)(nil))].(*CodecRegistry).Register(mediaType, c)
}

// Bind 向容器註冊一個依賴，s 省略時預設使用 SingletonScope。
func (ro *Router) Bind(name string, factory func() any, s ...scope.Scope) {
	ro.container.Register(name, factory, s...)
}

// Resolve 從容器取得指定名稱的依賴實體，供啟動時組裝使用。
func (ro *Router) Resolve(name string) any {
	return ro.container.Resolve(context.Background(), name)
}

func (ro *Router) register(h routing.HttpHandler) {
	ro.handlers = append(ro.handlers, h)
}

func (ro *Router) GET(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodGet, f)))
}

func (ro *Router) POST(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPost, f)))
}

func (ro *Router) PUT(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPut, f)))
}

func (ro *Router) DELETE(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodDelete, f)))
}

func (ro *Router) PATCH(path string, f HandlerFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPatch, f)))
}

// Run 啟動 HTTP server 並監聽指定 addr（如 ":8080"）。
func (ro *Router) Run(addr string) error {
	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, ro)
}

// ServeHTTP 實作 http.Handler 介面，是每個 HTTP request 的統一入口。
// Go 的 net/http 套件在收到請求時會自動呼叫此方法。
func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = ro.injectContext(r)
	ro.dispatch(w, r)
}

// injectContext 在 request 進入 handler 前，將所有必要資訊注入 context。
// 注入順序：
//  1. errorHandler：讓 routing 層可回傳 404/405
//  2. plugins（實作 RequestPreparer 者）：各 plugin 自行注入所需資料
//     （例如 CodecRegistry 注入 codec map、ExceptionMapperPlugin 注入 error 規則）
//  3. IoC Container：讓 handler 透過 Get[T] 解析依賴
//     （同時注入 HttpRequestScope 的 store，屬於 Container 的內部機制）
func (ro *Router) injectContext(r *http.Request) *http.Request {
	r = storeErrorHandler(r, ro.errorHandler)
	for _, p := range ro.plugins {
		if preparer, ok := p.(plugin.RequestPreparer); ok {
			r = preparer.PrepareRequest(r)
		}
	}
	if ro.container != nil {
		r = injectContainer(r, ro.container)
	}
	return r
}

// dispatch 依序讓每個 HttpHandler 嘗試處理請求。
// 若某個 handler 完整處理（Handled），立即返回。
// 否則持續追蹤最佳匹配結果（PathMatched > NotMatched），最終交給 handleRoutingError 決定 404 或 405。
func (ro *Router) dispatch(w http.ResponseWriter, r *http.Request) {
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
	handleRoutingError(w, r, ro.errorHandler, best)
}
