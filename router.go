package framework

import (
	"context"
	"log"
	"net/http"
	"reflect"

	"github.com/xchwan/simple-web-framework/builtin"
	"github.com/xchwan/simple-web-framework/plugin"
	"github.com/xchwan/simple-web-framework/routing"
	"github.com/xchwan/simple-web-framework/scope"
)

// HandlerFunc is a type alias for routing.HandlerFunc so callers do not need to import the routing package directly.
type HandlerFunc = routing.HandlerFunc

// MiddlewareFunc is a type alias for routing.MiddlewareFunc so callers do not need to import the routing package directly.
type MiddlewareFunc = routing.MiddlewareFunc

// PathParam retrieves a named path parameter from the request context.
func PathParam(r *http.Request, key string) string {
	return routing.PathParam(r, key)
}

// Router holds a set of HttpHandlers and tries each one in order for every incoming request.
type Router struct {
	handlers     []routing.HttpHandler
	errorHandler ErrorHandlerFunc
	container    *Container
	plugins      map[reflect.Type]any
}

// NewRouter creates and returns a new Router with the default error handler and
// built-in JSON and text/plain codecs pre-registered.
func NewRouter() *Router {
	r := &Router{
		errorHandler: builtin.DefaultErrorHandler,
		container:    NewContainer(),
		plugins:      make(map[reflect.Type]any),
	}
	cr := plugin.NewCodecRegistry()
	cr.Register("application/json", &builtin.JsonCodec{})
	cr.Register("text/plain", &builtin.TextCodec{})
	r.plugins[reflect.TypeOf(cr)] = cr
	return r
}

// SetErrorHandler overrides the default routing-error handler (404 / 405).
func (ro *Router) SetErrorHandler(f ErrorHandlerFunc) {
	ro.errorHandler = f
}

// AddPlugin installs a plugin, storing it in the plugins map keyed by its type.
// If p implements Installer, Install is called immediately with the current PluginContext.
// If p implements ContextInjector, Inject is called automatically on every request.
func (ro *Router) AddPlugin(p any) {
	ro.plugins[reflect.TypeOf(p)] = p
	if installer, ok := p.(plugin.Installer); ok {
		installer.Install(plugin.PluginContext(ro.plugins))
	}
}

// Bind registers a dependency with the container. Defaults to SingletonScope when no scope is provided.
func (ro *Router) Bind(name string, factory func() any, s ...scope.Scope) {
	ro.container.Register(name, factory, s...)
}

// Resolve retrieves a named dependency from the container. Intended for use during startup wiring.
func (ro *Router) Resolve(name string) any {
	return ro.container.Resolve(context.Background(), name)
}

func (ro *Router) register(h routing.HttpHandler) {
	ro.handlers = append(ro.handlers, h)
}

func (ro *Router) GET(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodGet, routing.ChainMiddlewares(f, m))))
}

func (ro *Router) POST(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPost, routing.ChainMiddlewares(f, m))))
}

func (ro *Router) PUT(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPut, routing.ChainMiddlewares(f, m))))
}

func (ro *Router) DELETE(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodDelete, routing.ChainMiddlewares(f, m))))
}

func (ro *Router) PATCH(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPatch, routing.ChainMiddlewares(f, m))))
}

// Run starts the HTTP server and listens on the given address (e.g. ":8080").
func (ro *Router) Run(addr string) error {
	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, ro)
}

// ServeHTTP implements http.Handler and is the single entry point for every HTTP request.
func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = ro.injectContext(r)
	ro.dispatch(w, r)
}

// injectContext enriches the request context before it reaches any handler.
// Injection order:
//  1. errorHandler — makes the 404/405 handler available to the routing layer
//  2. plugins (ContextInjector) — each plugin injects its own data (e.g. codec map, error rules)
//  3. IoC container — enables Get[T] dependency resolution inside handlers
func (ro *Router) injectContext(r *http.Request) *http.Request {
	r = storeErrorHandler(r, ro.errorHandler)
	for _, p := range ro.plugins {
		if injector, ok := p.(plugin.ContextInjector); ok {
			r = injector.Inject(r)
		}
	}
	if ro.container != nil {
		r = injectContainer(r, ro.container)
	}
	return r
}

// dispatch tries each registered HttpHandler in order.
// Returns as soon as one handler reports Handled.
// Otherwise tracks the best partial match (PathMatched > NotMatched)
// and delegates to handleRoutingError to produce a 404 or 405.
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
