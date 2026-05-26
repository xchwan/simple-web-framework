package framework

// Routes is implemented by any type that can register HTTP routes.
// Both Router and Group satisfy this interface, so module wiring functions
// can accept either without caring which one they receive.
//
//	func RegisterUserRoutes(r framework.Routes) {
//	    g := r.Group("/api/users", Auth)
//	    g.GET("",      h.List)
//	    g.POST("",     h.Create)
//	}
//
//	RegisterUserRoutes(router)       // top-level router
//	RegisterUserRoutes(adminGroup)   // already-prefixed group
type Routes interface {
	GET(path string, f HandlerFunc, m ...MiddlewareFunc)
	POST(path string, f HandlerFunc, m ...MiddlewareFunc)
	PUT(path string, f HandlerFunc, m ...MiddlewareFunc)
	PATCH(path string, f HandlerFunc, m ...MiddlewareFunc)
	DELETE(path string, f HandlerFunc, m ...MiddlewareFunc)
	Group(prefix string, m ...MiddlewareFunc) *Group
}

// Compile-time checks: both Router and Group must implement Routes.
var _ Routes = (*Router)(nil)
var _ Routes = (*Group)(nil)
