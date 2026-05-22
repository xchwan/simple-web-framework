# 🕸️ Simple Web Framework

A lightweight HTTP framework built on top of Go's standard library, demonstrating how to implement an extensible web framework using **IoC Container**, **Plugin System**, and **Codec Registry** design patterns.

```bash
go get github.com/xchwan/simple-web-framework
```

---

## 📋 Table of Contents

- [🚀 Quick Start](#-quick-start)
- [🗺️ Routing](#️-routing)
- [📥 Request Parsing](#-request-parsing)
- [📤 Response Serialization](#-response-serialization)
- [🚨 Error Handling](#-error-handling)
- [🧩 IoC Container & Dependency Injection](#-ioc-container--dependency-injection)
- [♻️ Scopes (Lifecycle)](#️-scopes-lifecycle)
- [🔌 Plugin System](#-plugin-system)
- [🗜️ Codec Extension](#️-codec-extension)
- [📦 Full Example](#-full-example)
- [🛠️ Development Commands](#️-development-commands)
- [👤 Author](#-author)

---

## 🚀 Quick Start

```go
package main

import (
    "net/http"

    framework "github.com/xchwan/simple-web-framework"
)

func main() {
    r := framework.NewRouter()

    r.GET("/hello", func(w http.ResponseWriter, r *http.Request) {
        framework.Respond(w, r, http.StatusOK, map[string]string{"message": "Hello, World!"})
    })

    r.Run(":8080")
}
```

---

## 🗺️ Routing

### Basic Routes

Supports `GET`, `POST`, `PUT`, `DELETE`, and `PATCH` HTTP methods.

```go
r := framework.NewRouter()

r.GET("/api/users", listUsers)
r.POST("/api/users", createUser)
r.PUT("/api/users/{id}", updateUser)
r.PATCH("/api/users/{id}", patchUser)
r.DELETE("/api/users/{id}", deleteUser)
```

### 🔖 Path Parameters

Declare path parameters using `{name}` syntax and retrieve them with `framework.PathParam`.

```go
r.GET("/api/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
    id := framework.PathParam(r, "userId")
    // id == "42" when the request path is /api/users/42
})
```

### 🔍 Query Parameters

Use the standard library directly to extract query strings.

```go
r.GET("/api/users", func(w http.ResponseWriter, r *http.Request) {
    keyword := r.URL.Query().Get("keyword")
    // GET /api/users?keyword=alice  =>  keyword == "alice"
})
```

### ⚠️ Routing Error Behavior

| Scenario | HTTP Status Code |
|----------|-----------------|
| Path not found | 404 Not Found |
| Path matched but method not allowed | 405 Method Not Allowed |

---

## 📥 Request Parsing

The framework provides two parsing functions that automatically select a Codec based on the `Content-Type` header.

### ParseRequest — Manual Error Handling

Returns the error on decode failure; the caller decides how to respond.

```go
r.POST("/api/users", func(w http.ResponseWriter, req *http.Request) {
    var body CreateUserRequest
    if err := framework.ParseRequest(req, &body); err != nil {
        framework.HandleError(w, req, err)
        return
    }
    // use body ...
})
```

### ParseOrRespond — Automatic Error Handling

On decode failure, automatically calls `HandleError` (ExceptionMapper → `ErrBadRequest` default 400 → fallback 500). The handler only needs to check whether to `return`.

```go
r.POST("/api/users/login", func(w http.ResponseWriter, req *http.Request) {
    var body LoginRequest
    if err := framework.ParseOrRespond(w, req, &body); err != nil {
        return  // response already written by the framework
    }
    // use body ...
})
```

Supported out of the box: `application/json` and `text/plain` (see [🗜️ Codec Extension](#️-codec-extension)).

---

## 📤 Response Serialization

`framework.Respond` automatically selects a Codec based on the `Accept` header, serializes the body, and sets the appropriate response headers.

```go
// 200 OK with JSON body
framework.Respond(w, r, http.StatusOK, body)

// 201 Created
framework.Respond(w, r, http.StatusCreated, body)

// 204 No Content (no body written)
framework.Respond(w, r, http.StatusNoContent, nil)
```

**Error response format**: The framework uses `framework.ErrorBody` as the unified JSON error structure.

```go
// {"message": "something went wrong"}
framework.Respond(w, r, http.StatusBadRequest, framework.Error("something went wrong"))
```

---

## 🚨 Error Handling

### HandleError — Convert Go Errors to HTTP Responses

Tries the following three layers in order:

1. ⚡ **ExceptionMapperPlugin custom rules** — pointer equality first (O(1)), then `errors.Is` traversal for wrapped errors
2. 🛡️ **Framework defaults** — `framework.ErrBadRequest` → 400 Bad Request (no extra setup needed)
3. 🔥 **Fallback** — 500 Internal Server Error

```go
r.POST("/api/users", func(w http.ResponseWriter, req *http.Request) {
    if err := userService.Register(body.Email, body.Name, body.Password); err != nil {
        framework.HandleError(w, req, err)  // automatically maps err → HTTP status
        return
    }
    framework.Respond(w, req, http.StatusCreated, nil)
})
```

### 🛡️ ErrBadRequest — Framework Default Sentinel

`framework.ErrBadRequest` is a framework-level sentinel representing a malformed request (e.g., JSON parse failure). `HandleError` automatically responds with 400 when it encounters this error — **no ExceptionMapperPlugin configuration required**.

`ParseOrRespond` returns `ErrBadRequest` on decode failure. You can also use it directly:

```go
if someFormatInvalid {
    framework.HandleError(w, r, framework.ErrBadRequest)
    return
}
```

### 🗂️ ExceptionMapperPlugin — Define Business Error Mappings

Install the plugin during router setup to map all domain errors to HTTP status codes in one place.

```go
import "github.com/xchwan/simple-web-framework/plugin"

router.AddPlugin(
    plugin.NewExceptionMapperPlugin().
        On(ErrEmailDuplicate,     http.StatusBadRequest,   "Duplicate email").
        On(ErrCredentialsInvalid, http.StatusBadRequest,   "Credentials invalid").
        On(ErrTokenInvalid,       http.StatusUnauthorized, "Can't authenticate who you are.").
        On(ErrForbidden,          http.StatusForbidden,    "Forbidden"),
)
```

### 🎨 Custom Default Error Handler

Override the default response format for routing-layer errors (404 / 405).

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

---

## 🧩 IoC Container & Dependency Injection

### Registering Dependencies

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

### Resolving Dependencies in Handlers

Use `framework.Get[T]` — a type-safe generic function — to retrieve dependencies from the request context.

```go
r.GET("/api/users", func(w http.ResponseWriter, req *http.Request) {
    svc := framework.Get[*UserService](req, "userService")
    users := svc.SearchUsers("")
    framework.Respond(w, req, http.StatusOK, users)
})
```

### 🏗️ Resolving at Startup

`router.Resolve` can retrieve Singleton dependencies during router setup (outside of a request) to wire up handlers at startup.

```go
router.Bind("userRepo",    func() any { return NewUserRepository() })
router.Bind("userHandler", func() any { return NewUserHandler() })

h := router.Resolve("userHandler").(*UserHandler)
router.GET("/api/users", h.List)
```

---

## ♻️ Scopes (Lifecycle)

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

---

## 🔌 Plugin System

Plugins extend the framework through two focused interfaces:

```go
// Installer is called once when AddPlugin is invoked — for startup initialization
// (e.g., registering a codec into CodecRegistry).
type Installer interface {
    Install(ctx PluginContext)
}

// ContextInjector is called on every incoming request to inject data into the request context.
type ContextInjector interface {
    Inject(r *http.Request) *http.Request
}
```

A plugin can implement one or both interfaces.

### Installing a Plugin

```go
router.AddPlugin(myPlugin)
```

- If the plugin implements `Installer` → `Install` is called immediately with the current `PluginContext`
- If the plugin implements `ContextInjector` → `Inject` is called automatically on every request

### PluginContext — Bridge Between Plugins

`PluginContext` is a `map[reflect.Type]any` passed to `Install`, giving each plugin access to all currently registered resources. This allows plugins to collaborate without the Router knowing about concrete types.

```go
// XmlCodec registers itself into CodecRegistry during Install
func (c *XmlCodec) Install(ctx plugin.PluginContext) {
    ctx[reflect.TypeOf((*plugin.CodecRegistry)(nil))].(*plugin.CodecRegistry).
        Register("application/xml", c)
}
```

### 📦 Built-in Plugins

| Plugin | Interface | Function | Default |
|--------|-----------|----------|---------|
| `CodecRegistry` | `ContextInjector` | JSON + text/plain serialization, injected per request | ✅ Auto-installed |
| `ExceptionMapperPlugin` | `ContextInjector` | Maps errors to HTTP status codes, injected per request | 🔧 Manual |
| `XmlCodec` | `Installer` | Registers `application/xml` support into CodecRegistry | 🔧 Manual |

---

## 🗜️ Codec Extension

### 🗂️ Enable XML Support

The framework ships with a built-in `XmlCodec`. Install it to handle `application/xml` requests and responses.

```go
import "github.com/xchwan/simple-web-framework/plugin"

router.AddPlugin(&plugin.XmlCodec{})
```

### Adding a Custom Media Type

Implement `plugin.Codec` and register it via `Installer`:

```go
import (
    "io"
    "reflect"

    "github.com/xchwan/simple-web-framework/plugin"
)

type MsgpackCodec struct{}

func (c *MsgpackCodec) Install(ctx plugin.PluginContext) {
    ctx[reflect.TypeOf((*plugin.CodecRegistry)(nil))].(*plugin.CodecRegistry).
        Register("application/msgpack", c)
}

func (c *MsgpackCodec) Encode(w io.Writer, v any) error {
    return msgpack.NewEncoder(w).Encode(v)
}

func (c *MsgpackCodec) Decode(r io.Reader, v any) error {
    return msgpack.NewDecoder(r).Decode(v)
}

// Install
router.AddPlugin(&MsgpackCodec{})
```

---

## 📦 Full Example

The following shows the complete wiring flow from [`simple-web-app`](https://github.com/xchwan/simple-web-app) — a demo user service built on top of this framework.

### 1. 🐛 Define Domain Errors

```go
var (
    ErrEmailDuplicate        = errors.New("email duplicate")
    ErrRegisterFormatInvalid = errors.New("register format invalid")
    ErrCredentialsInvalid    = errors.New("credentials invalid")
    ErrLoginFormatInvalid    = errors.New("login format invalid")
    ErrTokenInvalid          = errors.New("token invalid")
    ErrForbidden             = errors.New("forbidden")
    ErrNameFormatInvalid     = errors.New("name format invalid")
)
```

### 2. ✍️ Write Handlers

Handlers retrieve the service from the container via `framework.Get[T]` — no dependencies held directly.

```go
type UserHandler struct{}

func (h *UserHandler) service(r *http.Request) *UserService {
    return framework.Get[*UserService](r, "userService")
}

// Register: lets the service validate fields and return domain errors (manual flow)
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerRequest
    framework.ParseRequest(r, &req)  // zero-value on failure; service validates
    u, err := h.service(r).Register(req.Email, req.Name, req.Password)
    if err != nil {
        framework.HandleError(w, r, err)
        return
    }
    framework.Respond(w, r, http.StatusCreated, userResponse{ID: u.ID, Email: u.Email, Name: u.Name})
}

// Login: bad request body → automatic 400, no service call (auto flow)
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginRequest
    if err := framework.ParseOrRespond(w, r, &req); err != nil {
        return
    }
    u, err := h.service(r).Login(req.Email, req.Password)
    if err != nil {
        framework.HandleError(w, r, err)
        return
    }
    framework.Respond(w, r, http.StatusOK, loginResponse{ID: u.ID, Email: u.Email, Name: u.Name, Token: u.Token})
}
```

### 3. 🔧 Wire Up Routes

```go
func Register(router *framework.Router) {
    router.AddPlugin(
        plugin.NewExceptionMapperPlugin().
            On(ErrEmailDuplicate,        http.StatusBadRequest,   "Duplicate email").
            On(ErrRegisterFormatInvalid, http.StatusBadRequest,   "Registration's format incorrect.").
            On(ErrCredentialsInvalid,    http.StatusBadRequest,   "Credentials Invalid").
            On(ErrLoginFormatInvalid,    http.StatusBadRequest,   "Login's format incorrect.").
            On(ErrTokenInvalid,          http.StatusUnauthorized, "Can't authenticate who you are.").
            On(ErrForbidden,             http.StatusForbidden,    "Forbidden").
            On(ErrNameFormatInvalid,     http.StatusBadRequest,   "Name's format invalid."),
    )

    router.Bind("userRepo", func() any { return NewUserRepository() })
    router.Bind("userService", func() any {
        repo := router.Resolve("userRepo").(*UserRepository)
        return NewUserService(repo)
    }, scope.NewHttpRequestScope())
    router.Bind("userHandler", func() any { return NewUserHandler() })

    h := router.Resolve("userHandler").(*UserHandler)
    router.POST("/api/users",           h.Register)
    router.POST("/api/users/login",     h.Login)
    router.PATCH("/api/users/{userId}", h.UpdateName)
    router.GET("/api/users",            h.SearchUsers)
}
```

### 4. 🚀 Start the Server

```go
func main() {
    r := framework.NewRouter()
    user.Register(r)
    r.Run(":8080")
}
```

---

## 🛠️ Development Commands

All commands run inside Docker — no local Go installation required.

```bash
make all          # ✅ staticcheck + format + test + build (full CI pipeline)
make test         # 🧪 Run integration tests under ./test/...
make build        # 🏗️ Compile binary
make staticcheck  # 🔍 Static analysis
make format       # 🎨 gofmt
make tidy         # 📦 go mod tidy
make shell        # 🐚 Interactive container shell
make clean        # 🧹 Remove binary and build cache
```

---

## 👤 Author

**xchwan**

- GitHub: [@xchwan](https://github.com/xchwan)
- Email: qchwan@gmail.com

---

*Contributions are welcome! The `.claude/` directory and `CLAUDE.md` are checked in to help contributors get started with Claude Code without additional setup.*
