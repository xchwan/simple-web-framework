# 🔌 Plugin System

Plugins extend the framework through three focused interfaces, each firing at a different point in the lifecycle:

```go
// Installer is called once when AddPlugin is invoked — for startup initialization
// (e.g., registering a codec into CodecRegistry).
type Installer interface {
    Install(ctx PluginContext)
}

// RouteHook is called once per route at registration time, before the server starts.
// Useful for collecting route metadata (e.g. for documentation generation).
type RouteHook interface {
    RouteAdded(method, path string, f HandlerFunc)
}

// ContextInjector is called on every incoming request to inject data into the request context.
type ContextInjector interface {
    Inject(r *http.Request) *http.Request
}
```

A plugin can implement any combination of these interfaces.

## Lifecycle Overview

| Interface | When it fires | Use case |
|-----------|--------------|----------|
| `Installer` | Once at `AddPlugin` | Register codecs, set up shared resources |
| `RouteHook` | Once per route registration | Collect route metadata, generate docs |
| `ContextInjector` | Every incoming request | Inject per-request data into context |

## Installing a Plugin

```go
router.AddPlugin(myPlugin)
```

- If the plugin implements `Installer` → `Install` is called immediately with the current `PluginContext`
- If the plugin implements `RouteHook` → `RouteAdded` is called once for every route registered after this point
- If the plugin implements `ContextInjector` → `Inject` is called automatically on every request

## PluginContext — Bridge Between Plugins

`PluginContext` is a `map[reflect.Type]any` passed to `Install`, giving each plugin access to all currently registered resources. This allows plugins to collaborate without the Router knowing about concrete types.

```go
// XmlCodec registers itself into CodecRegistry during Install
func (c *XmlCodec) Install(ctx plugin.PluginContext) {
    ctx[reflect.TypeOf((*plugin.CodecRegistry)(nil))].(*plugin.CodecRegistry).
        Register("application/xml", c)
}
```

## Built-in Plugins

| Plugin | Interface | Function | Default |
|--------|-----------|----------|---------|
| `CodecRegistry` | `ContextInjector` | JSON + text/plain serialization, injected per request | ✅ Auto-installed |
| `ExceptionMapperPlugin` | `ContextInjector` | Maps errors to HTTP status codes, injected per request | 🔧 Manual |
| `XmlCodec` | `Installer` | Registers `application/xml` support into CodecRegistry | 🔧 Manual |
| `DocPlugin` | `RouteHook` | Collects route metadata, serves OpenAPI 3.0 + Swagger UI | 🔧 Manual |

## Plugins vs Hooks

**Plugins** extend framework capabilities — they add new codecs, map errors, inject request-scoped data.

**Hooks** observe framework behaviour — they log, collect metrics, trace. See [Hook System](hooks.md).
