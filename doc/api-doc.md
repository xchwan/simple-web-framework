# 📖 API Documentation (Swagger UI)

The framework can automatically generate interactive API documentation powered by [Swagger UI](https://swagger.io/tools/swagger-ui/). Documentation is collected **eagerly at registration time** — every route appears in the docs regardless of whether it has been called.

## Setup

```go
import "github.com/xchwan/simple-web-framework/plugin"

docs := plugin.NewDocPlugin()
router.AddPlugin(docs)

// Routes registered with Doc[Req, Resp] appear in the docs
router.POST("/api/users",        h.Create, plugin.Doc[CreateUserRequest, UserResponse](docs))
router.POST("/api/users/login",  h.Login,  plugin.Doc[LoginRequest, LoginResponse](docs))
router.GET("/api/users",         h.List,   plugin.Doc[plugin.NoBody, []UserResponse](docs))
router.DELETE("/api/users/{id}", h.Delete, plugin.Doc[plugin.NoBody, plugin.NoBody](docs))

// Routes without Doc are unaffected — they simply don't appear in the docs
router.GET("/health", h.Health)

// Serve the docs
router.GET("/docs",         docs.UIHandler())    // Swagger UI page
router.GET("/openapi.json", docs.SpecHandler())  // OpenAPI 3.0 JSON spec
```

Open `http://localhost:8080/docs` to see the interactive documentation.

## Doc[Req, Resp]

`plugin.Doc[Req, Resp](docs)` is the last argument to any route registration method. It:

1. Stores the request/response type metadata in `docs` keyed by the handler's function pointer
2. Returns the original handler **unchanged** — no runtime overhead on the request path
3. When the router registers the route, it calls `docs.OnRegister(method, path, f)` which matches the pointer and records method + path + schema

## NoBody

Use `plugin.NoBody` as a type parameter when a route has no request or response body:

```go
// GET has no request body
router.GET("/api/users", h.List, plugin.Doc[plugin.NoBody, []UserResponse](docs))

// DELETE has neither request nor response body
router.DELETE("/api/users/{id}", h.Delete, plugin.Doc[plugin.NoBody, plugin.NoBody](docs))
```

## How It Works

`DocPlugin` implements the `RouteHook` interface, which fires once per route at registration time:

```
router.POST("/api/users", h.Create, plugin.Doc[CreateUserRequest, UserResponse](docs))
  │
  ├── plugin.Doc[...]  stores { reqType, respType } in docs.pending, keyed by f's pointer
  │                    returns f unchanged
  │
  └── router.POST      calls docs.OnRegister("POST", "/api/users", f)
                         → matches pointer → records route doc → clears pending
```

This means all metadata is collected **before the server starts**, and `/openapi.json` is fully populated from the first request.

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /docs` | Swagger UI — interactive HTML page, loads from CDN |
| `GET /openapi.json` | OpenAPI 3.0 JSON spec |
