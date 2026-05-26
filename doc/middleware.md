# 🔗 Middleware Chain

Middlewares wrap a handler using the **Decorator pattern**, running left-to-right before (and optionally after) the handler.

```go
router.GET("/api/events", h.List, Auth, RateLimit)
// execution order: Auth → RateLimit → h.List
```

## Middleware Signature

A middleware has the signature `func(next HandlerFunc) HandlerFunc`:

```go
func Auth(next framework.HandlerFunc) framework.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !validate(token) {
            framework.HandleError(w, r, ErrTokenInvalid)
            return  // short-circuit: handler never called
        }
        next(w, r)
    }
}

func RateLimit(next framework.HandlerFunc) framework.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if exceeded() {
            framework.HandleError(w, r, ErrRateLimitExceeded)
            return
        }
        next(w, r)
    }
}
```

## Post-handler Logic

Middlewares can also run code **after** the handler by placing logic after `next(w, r)`:

```go
func Timing(next framework.HandlerFunc) framework.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next(w, r)
        log.Printf("%s %s took %v", r.Method, r.URL.Path, time.Since(start))
    }
}
```

## Middleware vs Hook

Use **middleware** when you need to intercept, modify, or short-circuit request behaviour.

Use **hooks** when you only need to observe what happened (logging, metrics, tracing) — see [Hook System](hooks.md).
