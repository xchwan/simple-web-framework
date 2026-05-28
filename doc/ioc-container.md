# 🧩 IoC Container

The container manages dependencies that need a **controlled lifecycle**. Not everything belongs in it:

- **Singleton** (DB, Redis, repositories) — construct directly. No `Bind` needed.
- **HttpRequestScope / Prototype** — use `Bind` + `Get[T]`. The container manages when instances are created and how long they live.

## Example

```go
import "github.com/xchwan/simple-web-framework/scope"

// Singletons — just construct directly, no Bind
db   := NewUserDB(dsn)
repo := NewUserRepo(db)
rdb  := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// HttpRequestScope — needs Bind so Get[T] can retrieve it inside handlers.
// repo and rdb are captured from the outer scope.
router.Bind("userService", func() any {
    return NewUserService(repo, rdb)
}, scope.NewHttpRequestScope())
```

## Resolving in Handlers

Use `framework.Get[T]` to retrieve a bound dependency from the request context:

```go
router.GET("/api/users", func(w http.ResponseWriter, r *http.Request) {
    svc := framework.Get[*UserService](r, "userService")
    framework.Respond(w, r, http.StatusOK, svc.ListUsers())
})
```

Because `userService` is `HttpRequestScope`, every `Get[T]` call within the same request returns the same instance — middleware and handler share one `UserService`.

## Scopes

| Scope | Instance lifetime | When to use |
|-------|------------------|-------------|
| 🔒 Singleton | App lifetime | DB, Redis, repositories — but just construct directly |
| 🌐 `HttpRequestScope` | Single HTTP request | Services with per-request state (e.g. current user, transaction) |
| 🆕 `PrototypeScope` | Per `Get[T]` call | Objects that must not be shared between callers |

## Common Mistake: Resolving Without Binding

Calling `Get[T]` for a name that was never registered panics immediately:

```
panic: dependency "userService" not found — register it with router.Bind("userService", func() any { return ... })
```

```go
// ✅ correct
router.Bind("userService", func() any { return NewUserService(repo) }, scope.NewHttpRequestScope())
svc := framework.Get[*UserService](r, "userService")

// ❌ panics — "userService" was never bound
svc := framework.Get[*UserService](r, "userService")
```
