package routing

import (
	"context"
	"net/http"
	"strings"
)

// PathHandler forwards a request to the wrapped handler only when the URL path matches the pattern.
// Patterns support {param} syntax, e.g. /api/users/{id}.
type PathHandler struct {
	segments []string
	wrapped  HttpHandler
}

// NewPathHandler creates a PathHandler that wraps the given handler.
func NewPathHandler(pattern string, wrapped HttpHandler) *PathHandler {
	return &PathHandler{
		segments: strings.Split(strings.Trim(pattern, "/"), "/"),
		wrapped:  wrapped,
	}
}

// Handle implements HttpHandler. Returns NotMatched when the path does not match.
// On match, extracts path parameters into the request context and delegates to the wrapped handler.
func (d *PathHandler) Handle(w http.ResponseWriter, r *http.Request) HandleResult {
	params, ok := matchPath(d.segments, r.URL.Path)
	if !ok {
		return NotMatched
	}
	r = r.WithContext(context.WithValue(r.Context(), pathParamsKey{}, params))
	return d.wrapped.Handle(w, r)
}

// matchPath compares pattern segments against the actual path and returns extracted path parameters.
func matchPath(segments []string, path string) (map[string]string, bool) {
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) != len(pathSegments) {
		return nil, false
	}
	params := map[string]string{}
	for i, seg := range segments {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			params[seg[1:len(seg)-1]] = pathSegments[i]
		} else if seg != pathSegments[i] {
			return nil, false
		}
	}
	return params, true
}

// pathParamsKey is the context key used to store and retrieve path parameters.
type pathParamsKey struct{}

// PathParam retrieves a named path parameter from the request context.
func PathParam(r *http.Request, key string) string {
	if params, ok := r.Context().Value(pathParamsKey{}).(map[string]string); ok {
		return params[key]
	}
	return ""
}
