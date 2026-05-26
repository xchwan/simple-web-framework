package plugin

import (
	"net/http"
	"reflect"

	"github.com/xchwan/simple-web-framework/routing"
)

// HandlerFunc is an alias so plugin code can reference the handler type without importing routing directly.
type HandlerFunc = routing.HandlerFunc

// PluginContext is a map of all currently registered resources, keyed by type.
// It is passed to Installer.Install so plugins can look up other resources (e.g. CodecRegistry)
// without the Router knowing about any concrete types.
type PluginContext map[reflect.Type]any

// Installer is implemented by plugins that need one-time initialisation at install time
// (e.g. registering a codec into CodecRegistry).
type Installer interface {
	Install(ctx PluginContext)
}

// ContextInjector is implemented by plugins that need to inject data into the request context
// on every incoming request.
type ContextInjector interface {
	Inject(r *http.Request) *http.Request
}

// RouteHook is implemented by plugins that need to observe route registrations.
// OnRegister is called once per route at registration time, before the server starts.
type RouteHook interface {
	OnRegister(method, path string, f HandlerFunc)
}
