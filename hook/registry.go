// Package hook provides observation points for framework lifecycle events.
// Unlike plugins, hooks observe framework behaviour without extending it.
package hook

import (
	"context"
	"net/http"
)

type hooksKey struct{}

// Hooks holds all registered hooks for each observation point.
type Hooks struct {
	onRequest []OnRequestFunc
	onRespond []OnRespondFunc
	onError   []OnErrorFunc
}

// Inject stores the registry in the request context.
func (reg *Hooks) Inject(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), hooksKey{}, reg))
}

// Load retrieves the registry from the request context.
func Load(r *http.Request) *Hooks {
	reg, _ := r.Context().Value(hooksKey{}).(*Hooks)
	return reg
}
