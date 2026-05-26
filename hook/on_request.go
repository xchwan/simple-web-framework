package hook

import "net/http"

// OnRequestFunc is called when an incoming request is received, before dispatch.
type OnRequestFunc func(r *http.Request)

// AddOnRequest registers a hook that fires on every incoming request.
func (reg *Hooks) AddOnRequest(f OnRequestFunc) {
	reg.onRequest = append(reg.onRequest, f)
}

// NotifyRequest calls all registered OnRequest hooks.
func (reg *Hooks) NotifyRequest(r *http.Request) {
	for _, f := range reg.onRequest {
		f(r)
	}
}
