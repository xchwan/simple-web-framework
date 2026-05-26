package framework

import "net/http"

// Group is a route builder that prepends a common path prefix and a set of
// shared middlewares to every route registered through it.
// Create one with Router.Group; the underlying Router is unaware of groups —
// each call to g.GET / g.POST / … expands directly into a top-level route.
type Group struct {
	router     *Router
	prefix     string
	middleware []MiddlewareFunc
}

// Group returns a new Group whose routes share the given path prefix and
// middlewares. The middlewares are prepended to any per-route middlewares,
// so execution order is: group middlewares → route middlewares → handler.
func (ro *Router) Group(prefix string, m ...MiddlewareFunc) *Group {
	return &Group{
		router:     ro,
		prefix:     prefix,
		middleware: m,
	}
}

func (g *Group) merged(m []MiddlewareFunc) []MiddlewareFunc {
	if len(g.middleware) == 0 {
		return m
	}
	out := make([]MiddlewareFunc, 0, len(g.middleware)+len(m))
	out = append(out, g.middleware...)
	out = append(out, m...)
	return out
}

func (g *Group) GET(path string, f HandlerFunc, m ...MiddlewareFunc) {
	g.router.GET(g.prefix+path, f, g.merged(m)...)
}

func (g *Group) POST(path string, f HandlerFunc, m ...MiddlewareFunc) {
	g.router.POST(g.prefix+path, f, g.merged(m)...)
}

func (g *Group) PUT(path string, f HandlerFunc, m ...MiddlewareFunc) {
	g.router.PUT(g.prefix+path, f, g.merged(m)...)
}

func (g *Group) PATCH(path string, f HandlerFunc, m ...MiddlewareFunc) {
	g.router.PATCH(g.prefix+path, f, g.merged(m)...)
}

func (g *Group) DELETE(path string, f HandlerFunc, m ...MiddlewareFunc) {
	g.router.DELETE(g.prefix+path, f, g.merged(m)...)
}

// Group creates a nested group that inherits this group's prefix and
// middlewares, then appends the given prefix and middlewares on top.
func (g *Group) Group(prefix string, m ...MiddlewareFunc) *Group {
	return &Group{
		router:     g.router,
		prefix:     g.prefix + prefix,
		middleware: g.merged(m),
	}
}

// HandleFunc registers a handler for an arbitrary HTTP method.
// Useful when the five convenience methods (GET/POST/PUT/PATCH/DELETE) are not enough.
func (g *Group) HandleFunc(method, path string, f HandlerFunc, m ...MiddlewareFunc) {
	switch method {
	case http.MethodGet:
		g.GET(path, f, m...)
	case http.MethodPost:
		g.POST(path, f, m...)
	case http.MethodPut:
		g.PUT(path, f, m...)
	case http.MethodPatch:
		g.PATCH(path, f, m...)
	case http.MethodDelete:
		g.DELETE(path, f, m...)
	}
}
