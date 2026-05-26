# 📦 Full Example

The following shows the complete wiring flow from [`simple-web-app`](https://github.com/xchwan/simple-web-app) — a demo user service built on top of this framework.

## 1. Define Domain Errors

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

## 2. Write Handlers

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

## 3. Wire Up Routes

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

## 4. Start the Server

```go
func main() {
    r := framework.NewRouter()
    user.Register(r)
    r.Run(":8080")
}
```
