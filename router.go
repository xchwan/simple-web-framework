package framework

import (
	"context"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/xchwan/simple-web-framework/builtin"
	"github.com/xchwan/simple-web-framework/hook"
	"github.com/xchwan/simple-web-framework/plugin"
	"github.com/xchwan/simple-web-framework/plugin/codec"
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
	handlers        []routing.HttpHandler
	errorHandler    ErrorHandlerFunc
	container       *Container
	plugins         map[reflect.Type]any
	hooks           *hook.Hooks
	shutdownTimeout time.Duration
}

// defaultShutdownTimeout is used when the caller has not set a custom timeout.
const defaultShutdownTimeout = 5 * time.Second

// NewRouter creates and returns a new Router with the default error handler and
// built-in JSON and text/plain codecs pre-registered.
func NewRouter() *Router {
	r := &Router{
		errorHandler:    builtin.DefaultErrorHandler,
		container:       NewContainer(),
		plugins:         make(map[reflect.Type]any),
		hooks:           &hook.Hooks{},
		shutdownTimeout: defaultShutdownTimeout,
	}
	cr := codec.NewCodecRegistry()
	cr.Register("application/json", &codec.JsonCodec{})
	cr.Register("text/plain", &codec.TextCodec{})
	r.plugins[reflect.TypeOf(cr)] = cr
	return r
}

// SetErrorHandler overrides the default routing-error handler (404 / 405).
func (ro *Router) SetErrorHandler(f ErrorHandlerFunc) {
	ro.errorHandler = f
}

// SetShutdownTimeout sets the maximum time Run will wait for in-flight requests
// to finish after a SIGINT or SIGTERM is received. Defaults to 5 seconds.
func (ro *Router) SetShutdownTimeout(d time.Duration) {
	ro.shutdownTimeout = d
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

// resolve retrieves a named dependency from the container. Internal use only.
func (ro *Router) resolve(name string) any {
	return ro.container.Resolve(context.Background(), name)
}

// OnRequest registers a hook that fires on every incoming request, before dispatch.
func (ro *Router) OnRequest(f hook.OnRequestFunc) {
	ro.hooks.AddOnRequest(f)
}

// OnRespond registers a hook that fires when Respond writes a successful response.
func (ro *Router) OnRespond(f hook.OnRespondFunc) {
	ro.hooks.AddOnRespond(f)
}

// OnError registers a hook that fires when HandleError writes an error response.
func (ro *Router) OnError(f hook.OnErrorFunc) {
	ro.hooks.AddOnError(f)
}

func (ro *Router) register(h routing.HttpHandler) {
	ro.handlers = append(ro.handlers, h)
}

// notifyRegister calls RouteAdded on every plugin that implements RouteHook.
func (ro *Router) notifyRegister(method, path string, f HandlerFunc) {
	for _, p := range ro.plugins {
		if rh, ok := p.(plugin.RouteHook); ok {
			rh.RouteAdded(method, path, f)
		}
	}
}

func (ro *Router) GET(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.notifyRegister(http.MethodGet, path, f)
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodGet, routing.ChainMiddlewares(f, m))))
}

func (ro *Router) POST(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.notifyRegister(http.MethodPost, path, f)
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPost, routing.ChainMiddlewares(f, m))))
}

func (ro *Router) PUT(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.notifyRegister(http.MethodPut, path, f)
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPut, routing.ChainMiddlewares(f, m))))
}

func (ro *Router) DELETE(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.notifyRegister(http.MethodDelete, path, f)
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodDelete, routing.ChainMiddlewares(f, m))))
}

func (ro *Router) PATCH(path string, f HandlerFunc, m ...MiddlewareFunc) {
	ro.notifyRegister(http.MethodPatch, path, f)
	ro.register(routing.NewPathHandler(path, routing.NewMethodHandler(http.MethodPatch, routing.ChainMiddlewares(f, m))))
}

// Run starts the HTTP server and listens on the given address (e.g. ":8080").
// It blocks until ctx is cancelled, then gracefully shuts down: new requests
// are rejected and in-flight requests are given up to SetShutdownTimeout
// (default 5s) to complete before the server exits.
//
// The caller is responsible for cancelling ctx (e.g. via signal.NotifyContext),
// which allows coordinating shutdown order with other components such as
// message consumers:
//
//	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
//	defer stop()
//
//	go consumer.Run(ctx)
//	router.Run(ctx, ":8080")  // blocks until ctx is cancelled
func (ro *Router) Run(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: ro}

	go func() {
		log.Printf("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("Shutting down (timeout: %s)…", ro.shutdownTimeout)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ro.shutdownTimeout)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

// ServeHTTP implements http.Handler and is the single entry point for every HTTP request.
func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = ro.injectContext(r)
	hook.Load(r).NotifyRequest(r)
	ro.dispatch(w, r)
}

// injectContext enriches the request context before it reaches any handler.
// Injection order:
//  1. errorHandler — makes the 404/405 handler available to the routing layer
//  2. plugins (ContextInjector) — each plugin injects its own data (e.g. codec map, error rules)
//  3. IoC container — enables Get[T] dependency resolution inside handlers
//  4. hook registry — makes OnRequest/OnRespond/OnError hooks available to Respond and HandleError
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
	r = ro.hooks.Inject(r)
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
