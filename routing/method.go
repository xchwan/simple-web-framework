package routing

import "net/http"

// MethodHandler forwards a request to the wrapped handler only when the HTTP method matches.
type MethodHandler struct {
	method  string
	wrapped HttpHandler
}

// NewMethodHandler creates a MethodHandler that wraps the given handler.
func NewMethodHandler(method string, wrapped HttpHandler) *MethodHandler {
	return &MethodHandler{method: method, wrapped: wrapped}
}

// Handle implements HttpHandler. Returns PathMatched when the method does not match,
// or delegates to the wrapped handler when it does.
func (d *MethodHandler) Handle(w http.ResponseWriter, r *http.Request) HandleResult {
	if r.Method != d.method {
		return PathMatched
	}
	return d.wrapped.Handle(w, r)
}
