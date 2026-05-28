# 📖 API Documentation (Swagger UI)

The framework can automatically generate interactive API documentation powered by [Swagger UI](https://swagger.io/tools/swagger-ui/). Documentation is collected **eagerly at registration time** — every route appears in the docs regardless of whether it has been called.

## Setup

```go
import "github.com/xchwan/simple-web-framework/plugin/apidoc"

docs := apidoc.NewDocPlugin()
router.AddPlugin(docs)

// Routes registered with Doc[Req, Resp] appear in the docs
router.POST("/api/users",        h.Create, apidoc.Doc[CreateUserRequest, UserResponse](h.Create))
router.POST("/api/users/login",  h.Login,  apidoc.Doc[LoginRequest, LoginResponse](h.Login))
router.GET("/api/users",         h.List,   apidoc.Doc[apidoc.NoBody, []UserResponse](h.List))
router.DELETE("/api/users/{id}", h.Delete, apidoc.Doc[apidoc.NoBody, apidoc.NoBody](h.Delete))

// Routes without Doc are unaffected — they simply don't appear in the docs
router.GET("/health", h.Health)

// Serve the docs
router.GET("/docs",         docs.UIHandler())    // Swagger UI page
router.GET("/openapi.json", docs.SpecHandler())  // OpenAPI 3.0 JSON spec
```

Open `http://localhost:8080/docs` to see the interactive documentation.

## Adding Descriptions

### Plain string — shorthand for Summary

```go
router.POST("/api/users", h.Create,
    apidoc.Doc[CreateUserRequest, UserResponse](h.Create, "Register a new user"))
```

### DocOption — explicit and composable

```go
router.POST("/api/users", h.Create,
    apidoc.Doc[CreateUserRequest, UserResponse](h.Create,
        apidoc.Summary("Register a new user"),
        apidoc.Description("Creates a new account. The email address must be unique."),
        apidoc.Tags("users"),
    ))
```

| Option | Swagger UI | OpenAPI field |
|--------|-----------|---------------|
| `Summary(s)` | Endpoint title | `summary` |
| `Description(s)` | Expanded detail text | `description` |
| `Tags(t...)` | Section grouping | `tags` |

## NoBody

Use `apidoc.NoBody` as a type parameter when a route has no request or response body:

```go
// GET has no request body
router.GET("/api/users", h.List, apidoc.Doc[apidoc.NoBody, []UserResponse](h.List))

// DELETE has neither request nor response body
router.DELETE("/api/users/{id}", h.Delete, apidoc.Doc[apidoc.NoBody, apidoc.NoBody](h.Delete))
```

## How It Works

`DocPlugin` implements the `RouteHook` interface, which fires once per route at registration time:

```
router.POST("/api/users", h.Create, apidoc.Doc[CreateUserRequest, UserResponse](h.Create))
  │
  ├── apidoc.Doc[...]  stores { reqType, respType, summary, … } in apidoc.pending (package-level),
  │                    keyed by h.Create's function pointer; returns h.Create unchanged
  │
  └── router.POST      calls docs.RouteAdded("POST", "/api/users", h.Create)
                         → matches pointer in apidoc.pending → records route doc → clears pending
```

All metadata is collected **before the server starts**, so `/openapi.json` is fully populated from the very first request.

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /docs` | Swagger UI — interactive HTML page, loads from CDN |
| `GET /openapi.json` | OpenAPI 3.0 JSON spec |

## Common Mistake: Forgetting `AddPlugin`

If `apidoc.Doc[]` is called but `router.AddPlugin(docs)` is omitted, routes work normally but nothing appears in the docs. When `/openapi.json` is first requested, a warning is printed:

```
[apidoc] warning: Doc[] was called but DocPlugin is not registered — call router.AddPlugin(docs)
```

Make sure `AddPlugin` is called **before** any route registration:

```go
docs := apidoc.NewDocPlugin()
router.AddPlugin(docs)          // must come first

router.POST("/api/users", h.Create, apidoc.Doc[CreateUserRequest, UserResponse](h.Create))
```
