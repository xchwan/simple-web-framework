# 🚨 Error Handling

## HandleError — Convert Go Errors to HTTP Responses

`HandleError` translates a Go error into an HTTP response. It tries three layers in order:

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

## ErrBadRequest — Framework Default Sentinel

`framework.ErrBadRequest` represents a malformed request (e.g. JSON parse failure). `HandleError` automatically responds with 400 when it encounters this error — **no ExceptionMapperPlugin configuration required**.

`ParseOrRespond` returns `ErrBadRequest` on decode failure. You can also use it directly:

```go
if someFormatInvalid {
    framework.HandleError(w, r, framework.ErrBadRequest)
    return
}
```

## ExceptionMapperPlugin — Define Business Error Mappings

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

## Custom Default Error Handler

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
