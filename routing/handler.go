// Package routing provides the core interfaces and components for HTTP routing.
package routing

import "net/http"

// HandleResult indicates how a handler responded to a request.
type HandleResult int

const (
	// NotMatched means this handler did not match at all; the Router continues to the next.
	NotMatched HandleResult = iota
	// PathMatched means the path matched but the HTTP method did not; the Router should return 405.
	PathMatched HandleResult = iota
	// Handled means the request was fully processed; the Router stops immediately.
	Handled HandleResult = iota
)

// HttpHandler is the interface implemented by all routing components in the framework.
type HttpHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) HandleResult
}

// HandlerFunc is a function adapter that satisfies the HttpHandler interface.
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Handle calls the function itself and returns Handled.
func (f HandlerFunc) Handle(w http.ResponseWriter, r *http.Request) HandleResult {
	f(w, r)
	return Handled
}
