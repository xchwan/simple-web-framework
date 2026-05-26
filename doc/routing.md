# 🗺️ Routing

## Basic Routes

Supports `GET`, `POST`, `PUT`, `DELETE`, and `PATCH` HTTP methods.

```go
r := framework.NewRouter()

r.GET("/api/users", listUsers)
r.POST("/api/users", createUser)
r.PUT("/api/users/{id}", updateUser)
r.PATCH("/api/users/{id}", patchUser)
r.DELETE("/api/users/{id}", deleteUser)
```

## Path Parameters

Declare path parameters using `{name}` syntax and retrieve them with `framework.PathParam`.

```go
r.GET("/api/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
    id := framework.PathParam(r, "userId")
    // id == "42" when the request path is /api/users/42
})
```

## Query Parameters

Use the standard library directly to extract query strings.

```go
r.GET("/api/users", func(w http.ResponseWriter, r *http.Request) {
    keyword := r.URL.Query().Get("keyword")
    // GET /api/users?keyword=alice  =>  keyword == "alice"
})
```

## Route Grouping

Use `router.Group` to share a common path prefix and middlewares across multiple routes.

```go
// All routes under /api/users, all protected by Auth
g := router.Group("/api/users", Auth)

g.GET("",      h.List)
g.POST("",     h.Create)
g.GET("/{id}", h.Get)
g.PATCH("/{id}", h.Update)
g.DELETE("/{id}", h.Delete)
```

Per-route middlewares can still be added alongside the group-level ones. Execution order is always: **group middlewares → route middlewares → handler**.

```go
g.GET("/export", h.Export, RateLimit)
// order: Auth → RateLimit → h.Export
```

### Nested Groups

Groups can be nested — each child inherits the parent's prefix and middlewares.

```go
api := router.Group("/api")
users := api.Group("/users", Auth)
users.GET("", h.List)         // GET /api/users

admin := api.Group("/admin", Auth, AdminOnly)
admin.GET("/stats", h.Stats)  // GET /api/admin/stats
```

## Routing Error Behavior

| Scenario | HTTP Status Code |
|----------|-----------------|
| Path not found | 404 Not Found |
| Path matched but method not allowed | 405 Method Not Allowed |

You can override the default format with a custom error handler:

```go
router.SetErrorHandler(func(w http.ResponseWriter, r *http.Request, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(map[string]string{
        "error": http.StatusText(statusCode),
        "path":  r.URL.Path,
    })
})
```
