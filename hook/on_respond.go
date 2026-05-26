package hook

import "net/http"

// OnRespondFunc is called when Respond writes a successful response.
type OnRespondFunc func(r *http.Request, statusCode int)

// AddOnRespond registers a hook that fires when Respond is called.
func (reg *Hooks) AddOnRespond(f OnRespondFunc) {
	reg.onRespond = append(reg.onRespond, f)
}

// NotifyRespond calls all registered OnRespond hooks.
func (reg *Hooks) NotifyRespond(r *http.Request, statusCode int) {
	for _, f := range reg.onRespond {
		f(r, statusCode)
	}
}
