# 🧩 IoC Container

## Registering Dependencies

Use `router.Bind` to register a named dependency. `Bind` only records the factory function — the function does **not** run until someone calls `Get[T]` for that name.

The default scope is **Singleton** — omit the third argument and you get one shared instance for the lifetime of the app.

For Singleton dependencies (DB connections, repositories), it is simpler to construct them directly and capture them in the factory closure rather than going through the container:

```go
import "github.com/xchwan/simple-web-framework/scope"

// Singletons — construct directly
db   := NewUserDB(dsn)
repo := NewUserRepo(db)
rdb  := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// 🌐 HttpRequestScope — a fresh userService per request.
// repo and rdb are captured from the outer scope (no Bind needed for them).
router.Bind("userService", func() any {
    return NewUserService(repo, rdb)
}, scope.NewHttpRequestScope())
```

## Resolving Dependencies in Handlers

Use `framework.Get[T]` inside a handler to retrieve a dependency from the request context:

```go
router.GET("/api/users", func(w http.ResponseWriter, r *http.Request) {
    svc := framework.Get[*UserService](r, "userService")
    users := svc.ListUsers()
    framework.Respond(w, r, http.StatusOK, users)
})
```

Because `userService` is `HttpRequestScope`, the same instance is returned for every `Get[T]` call within the same request.

## Scopes (Lifecycle)

| Scope | Description | Constructor |
|-------|-------------|-------------|
| 🔒 `SingletonScope` | One instance for the entire application | omit (default) |
| 🌐 `HttpRequestScope` | One instance shared within a single HTTP request | `scope.NewHttpRequestScope()` |
| 🆕 `PrototypeScope` | New instance on every `Get[T]` call | `scope.NewPrototypeScope()` |

**Choosing a scope:**
- Use **Singleton** (or just construct directly) for DB connections, Redis clients, repositories
- Use **HttpRequestScope** for service-layer objects that carry per-request state
- Use **Prototype** when each caller needs its own isolated instance (buffers, parsers)

## Common Mistake: Resolving Without Binding

Calling `Get[T]` for a name that was never registered panics immediately with a clear message:

```
panic: dependency "userService" not found — register it with router.Bind("userService", func() any { return ... })
```

Make sure every name you resolve has a corresponding `Bind`:

```go
// ✅ correct
router.Bind("userService", func() any { return NewUserService(repo) }, scope.NewHttpRequestScope())
svc := framework.Get[*UserService](r, "userService")

// ❌ panics — "userService" was never bound
svc := framework.Get[*UserService](r, "userService")
```
