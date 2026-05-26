# 🪝 Hook System

Hooks are **fire-and-forget observers** for framework lifecycle events. Unlike middleware, they cannot intercept or modify the request — they can only observe.

## Available Hooks

| Hook | Fires when | Signature |
|------|-----------|-----------|
| `OnRequest` | Every incoming request, before dispatch | `func(r *http.Request)` |
| `OnRespond` | `Respond` writes a successful response | `func(r *http.Request, statusCode int)` |
| `OnError` | `HandleError` writes an error response | `func(r *http.Request, err error)` |

## Hooks vs Middleware

| | Middleware | Hook |
|---|---|---|
| Intercept / short-circuit request | ✅ | ❌ |
| Modify response | ✅ | ❌ |
| Observe status code | 🔶 needs ResponseWriter wrapper | ✅ built-in |
| Logging, metrics, tracing | 🔶 complex | ✅ simple |

## Examples

### Logging

```go
router.OnRequest(func(r *http.Request) {
    log.Printf("→ %s %s", r.Method, r.URL.Path)
})
router.OnRespond(func(r *http.Request, statusCode int) {
    log.Printf("← %d %s %s", statusCode, r.Method, r.URL.Path)
})
```

### Metrics

```go
router.OnRespond(func(r *http.Request, statusCode int) {
    metrics.Inc("http.response", statusCode)
})
```

### Error Tracking

```go
router.OnError(func(r *http.Request, err error) {
    if !errors.Is(err, framework.ErrBadRequest) {
        sentry.CaptureException(err)
    }
})
```

### Distributed Tracing

```go
router.OnRequest(func(r *http.Request) {
    trace.Start(r.Context(), r.URL.Path)
})
router.OnRespond(func(r *http.Request, statusCode int) {
    trace.End(r.Context(), statusCode)
})
```

Multiple hooks of the same type can be registered — all of them fire in registration order.
