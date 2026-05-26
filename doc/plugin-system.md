# 🔌 Plugin System

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

## Installing a Plugin

```go
router.AddPlugin(myPlugin)
```

- If the plugin implements `Installer` → `Install` is called immediately with the current `PluginContext`
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

## Plugins vs Hooks

**Plugins** extend framework capabilities — they add new codecs, map errors, inject request-scoped data.

**Hooks** observe framework behaviour — they log, collect metrics, trace. See [Hook System](hooks.md).
