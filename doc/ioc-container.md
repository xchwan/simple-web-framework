# 🧩 IoC Container

## Registering Dependencies

Use `router.Bind` to register a named dependency. `Bind` only records the factory function — the function does **not** run until someone calls `Resolve` or `Get[T]` for that name.

```go
import "github.com/xchwan/simple-web-framework/scope"

// 🔒 Singleton (default) — created once, shared for the lifetime of the app
router.Bind("userRepo", func() any {
    return NewUserDB(db)
})
router.Bind("redis", func() any {
    return redis.NewClient(&redis.Options{Addr: "localhost:6379"})
})

// 🌐 HttpRequestScope — a fresh userService is created for every request.
// Inside the factory, Resolve pulls userRepo and redis from the singleton pool,
// so DB and Redis connections are still shared — only the service layer is per-request.
router.Bind("userService", func() any {
    repo := router.Resolve("userRepo").(*UserDB)
    rdb  := router.Resolve("redis").(*redis.Client)
    return NewUserService(repo, rdb)
}, scope.NewHttpRequestScope())
```

## Resolving Dependencies in Handlers

Use `framework.Get[T]` inside a handler to retrieve a dependency from the request context.

```go
router.GET("/api/users", func(w http.ResponseWriter, r *http.Request) {
    svc := framework.Get[*UserService](r, "userService")
    users := svc.ListUsers()
    framework.Respond(w, r, http.StatusOK, users)
})
```

`Get[T]` resolves `userService` for this request. Because it is `HttpRequestScope`, the same instance is returned for every `Get` call within the same request — two handlers in the same request pipeline share one `UserService`.

## Resolving at Startup

`router.Resolve` retrieves a Singleton outside of a request — useful for wiring handlers at startup.

```go
router.Bind("userRepo",    func() any { return NewUserDB(db) })
router.Bind("userHandler", func() any {
    repo := router.Resolve("userRepo").(*UserDB)
    return NewUserHandler(repo)
})

h := router.Resolve("userHandler").(*UserHandler)
router.GET("/api/users", h.List)
router.POST("/api/users", h.Create)
```

## Scopes (Lifecycle)

| Scope | Description | Constructor |
|-------|-------------|-------------|
| 🔒 `SingletonScope` (default) | One instance for the entire application | `scope.NewSingletonScope()` |
| 🌐 `HttpRequestScope` | One instance shared within a single HTTP request | `scope.NewHttpRequestScope()` |
| 🆕 `PrototypeScope` | New instance on every `Resolve` call | `scope.NewPrototypeScope()` |

**Choosing a scope:**
- Use **Singleton** for stateless or connection-holding objects (DB, Redis, repositories)
- Use **HttpRequestScope** for objects that carry per-request state (services, unit-of-work)
- Use **Prototype** when each caller needs its own isolated instance (buffers, parsers)
