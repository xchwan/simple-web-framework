package routing

// MiddlewareFunc wraps a HandlerFunc with additional behaviour (Decorator pattern).
// Middlewares are applied left-to-right: the first one in the list runs first.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// ChainMiddlewares wraps f with the given middlewares, outermost first.
func ChainMiddlewares(f HandlerFunc, middlewares []MiddlewareFunc) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		f = middlewares[i](f)
	}
	return f
}
