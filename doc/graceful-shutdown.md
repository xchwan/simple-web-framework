# 🛑 Graceful Shutdown

## The Problem

Without graceful shutdown, sending `Ctrl+C` or a `SIGTERM` (e.g. from Kubernetes rolling update) kills the process immediately. Any request that is mid-flight is forcibly cut off — the client receives a connection reset instead of a proper response.

This matters in practice:

- A user is checking out and the request is interrupted → they see an error page
- A request is mid-way through writing to the database → partial write, potential data corruption
- Kubernetes sends `SIGTERM` to the old pod during a rolling deploy → requests hitting that pod at that moment all fail

## How It Works

`Router.Run` accepts a `context.Context`. When the context is cancelled, it:

1. Stops accepting new requests
2. Waits for in-flight requests to finish (up to the configured timeout)
3. Returns

Signal handling lives in `main.go`, giving you full control over shutdown order:

```go
import (
    "context"
    "os/signal"
    "syscall"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    router := framework.NewRouter()
    // ... register routes ...
    router.Run(ctx, ":8080")
}
```

## Coordinating Multiple Components

Because `Run` takes an external `ctx`, you can coordinate shutdown order with other long-running components — for example, stopping a Kafka consumer only after the HTTP server has drained:

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

go consumer.Run(ctx)      // consumer stops when ctx is cancelled
router.Run(ctx, ":8080")  // HTTP server drains, then returns
// after Run returns, consumer is also stopped
```

## Setting the Timeout

The default drain timeout is 5 seconds. Override it before calling `Run`:

```go
router.SetShutdownTimeout(10 * time.Second)
router.Run(ctx, ":8080")
```

If in-flight requests do not finish within the timeout, the server exits anyway — the timeout is a safety net, not a guarantee.
