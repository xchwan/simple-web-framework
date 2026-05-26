package hook

import "net/http"

// OnErrorFunc is called when HandleError writes an error response.
type OnErrorFunc func(r *http.Request, err error)

// AddOnError registers a hook that fires when HandleError is called.
func (reg *Hooks) AddOnError(f OnErrorFunc) {
	reg.onError = append(reg.onError, f)
}

// NotifyError calls all registered OnError hooks.
func (reg *Hooks) NotifyError(r *http.Request, err error) {
	for _, f := range reg.onError {
		f(r, err)
	}
}
