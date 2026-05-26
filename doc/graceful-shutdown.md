# 🛑 Graceful Shutdown

## The Problem

Without graceful shutdown, sending `Ctrl+C` or a `SIGTERM` (e.g. from Kubernetes rolling update) kills the process immediately. Any request that is mid-flight is forcibly cut off — the client receives a connection reset instead of a proper response.

This matters in practice:

- A user is checking out and the request is interrupted → they see an error page
- A request is mid-way through writing to the database → partial write, potential data corruption
- Kubernetes sends `SIGTERM` to the old pod during a rolling deploy → requests hitting that pod at that moment all fail

## How It Works

`Router.Run` handles this automatically. When `SIGINT` or `SIGTERM` is received:

1. Stop accepting new requests
2. Wait for in-flight requests to finish (up to the configured timeout)
3. Exit cleanly

```
收到 SIGTERM
  → 停止接受新 request
  → 等現有 request 跑完（預設最多 5 秒）
  → 正常退出
```

No code changes are needed in handlers — `r.Run(":8080")` already does this.

## Setting the Timeout

The default timeout is 5 seconds. Override it with `SetShutdownTimeout` before calling `Run`:

```go
router := framework.NewRouter()
router.SetShutdownTimeout(10 * time.Second)
router.Run(":8080")
```

If in-flight requests do not finish within the timeout, the server exits anyway — the timeout is a safety net, not a guarantee.
