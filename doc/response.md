# 📤 Response Serialization

`framework.Respond` automatically selects a Codec based on the `Accept` header, serializes the body, and sets the appropriate response headers.

```go
// 200 OK with JSON body
framework.Respond(w, r, http.StatusOK, body)

// 201 Created
framework.Respond(w, r, http.StatusCreated, body)

// 204 No Content (no body written)
framework.Respond(w, r, http.StatusNoContent, nil)
```

## Error Response Format

The framework uses `framework.ErrorBody` as the unified JSON error structure.

```go
// {"message": "something went wrong"}
framework.Respond(w, r, http.StatusBadRequest, framework.Error("something went wrong"))
```

## OnRespond Hook

To observe response status codes (e.g. for logging or metrics), register an `OnRespond` hook instead of wrapping `ResponseWriter`:

```go
router.OnRespond(func(r *http.Request, statusCode int) {
    log.Printf("← %d %s %s", statusCode, r.Method, r.URL.Path)
})
```

See [Hook System](hooks.md) for details.
