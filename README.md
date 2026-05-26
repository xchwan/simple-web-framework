# 🕸️ Simple Web Framework

A lightweight HTTP framework built on top of Go's standard library, demonstrating how to implement an extensible web framework using **IoC Container**, **Plugin System**, and **Codec Registry** design patterns.

```bash
go get github.com/xchwan/simple-web-framework
```

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

## 📚 Documentation

| Topic | Description |
|-------|-------------|
| [🗺️ Routing](doc/routing.md) | Path params, query strings, HTTP methods, route grouping, `Routes` interface |
| [📥 Request Parsing](doc/request-parsing.md) | `ParseRequest` (manual) and `ParseOrRespond` (auto) |
| [📤 Response Serialization](doc/response.md) | Content-negotiation, error body format |
| [🚨 Error Handling](doc/error-handling.md) | `HandleError`, `ErrBadRequest`, `ExceptionMapperPlugin` |
| [🧩 IoC Container](doc/ioc-container.md) | `Bind`, `Resolve`, `Get[T]`, lifecycle scopes |
| [🔌 Plugin System](doc/plugin-system.md) | `Installer`, `ContextInjector`, `PluginContext` |
| [🗜️ Codec Extension](doc/codec-extension.md) | XML support, custom media types |
| [🔗 Middleware Chain](doc/middleware.md) | Decorator pattern, pre/post handler logic |
| [🪝 Hook System](doc/hooks.md) | `OnRequest`, `OnRespond`, `OnError` observers |
| [📦 Full Example](doc/full-example.md) | End-to-end user service wiring |
| [🎨 Design Patterns](doc/design-patterns.md) | Factory Method, Decorator, Chain of Responsibility, … |

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
