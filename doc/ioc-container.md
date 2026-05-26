# 🧩 IoC Container

## Registering Dependencies

Use `router.Bind` to register with the container. Defaults to `SingletonScope` when no scope is provided.

```go
// 🔒 Singleton — one instance shared across the entire application
router.Bind("userRepo", func() any {
    return NewUserRepository()
})

// 🌐 Explicit scope
router.Bind("userService", func() any {
    repo := router.Resolve("userRepo").(*UserRepository)
    return NewUserService(repo)
}, scope.NewHttpRequestScope())
```

## Resolving Dependencies in Handlers

Use `framework.Get[T]` — a type-safe generic function — to retrieve dependencies from the request context.

```go
r.GET("/api/users", func(w http.ResponseWriter, req *http.Request) {
    svc := framework.Get[*UserService](req, "userService")
    users := svc.SearchUsers("")
    framework.Respond(w, req, http.StatusOK, users)
})
```

## Resolving at Startup

`router.Resolve` can retrieve Singleton dependencies during router setup (outside of a request) to wire up handlers at startup.

```go
router.Bind("userRepo",    func() any { return NewUserRepository() })
router.Bind("userHandler", func() any { return NewUserHandler() })

h := router.Resolve("userHandler").(*UserHandler)
router.GET("/api/users", h.List)
```

## Scopes (Lifecycle)

| Scope | Description | Constructor |
|-------|-------------|-------------|
| 🔒 `SingletonScope` (default) | One instance for the entire application | `scope.NewSingletonScope()` |
| 🆕 `PrototypeScope` | New instance on every `Resolve` call | `scope.NewPrototypeScope()` |
| 🌐 `HttpRequestScope` | One instance shared within a single HTTP request | `scope.NewHttpRequestScope()` |

```go
import "github.com/xchwan/simple-web-framework/scope"

// Share one service instance per request
router.Bind("userService", func() any {
    return NewUserService()
}, scope.NewHttpRequestScope())

// Fresh instance on every resolve
router.Bind("tempBuffer", func() any {
    return &bytes.Buffer{}
}, scope.NewPrototypeScope())
```
