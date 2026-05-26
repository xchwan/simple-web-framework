# 📥 Request Parsing

The framework provides two parsing functions that automatically select a Codec based on the `Content-Type` header.

## ParseRequest — Manual Error Handling

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

## ParseOrRespond — Automatic Error Handling

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

## Supported Content Types

Out of the box: `application/json` and `text/plain`.

Additional types (e.g. `application/xml`, `application/msgpack`) can be added via the [Codec Extension](codec-extension.md) mechanism.
