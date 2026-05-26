# 🎨 Design Patterns

This framework is intentionally built around well-known design patterns. Here is a map of where each pattern appears and why it was chosen.

## Factory Method

**Where:** `router.Bind(name, func() any { ... })`

Each dependency is registered with a factory function — the Factory Method. The caller defines *how* to create the object; the container decides *when* to call it based on the configured scope. This keeps construction logic close to the dependency definition while letting the framework control the lifecycle.

## Singleton / Prototype

**Where:** `scope.SingletonScope`, `scope.PrototypeScope`, `scope.HttpRequestScope`

Three lifecycle scopes sit on top of the Factory Method layer. `SingletonScope` (default) calls the factory once and caches the result for the application lifetime; `PrototypeScope` calls it on every `Resolve`; `HttpRequestScope` calls it once per HTTP request and caches it on the request context.

## Chain of Responsibility

**Where:** `Router.dispatch` and `Router.injectContext`

The pattern appears in two forms:

- **Classic (stop-on-match)** — `Router.dispatch` tries each registered `HttpHandler` in order. The first handler that fully processes the request short-circuits the chain. If none match, the best partial result determines the 404 / 405 response.
- **Exhaustive variant (all-run)** — `Router.injectContext` passes the request through every `ContextInjector` plugin in sequence. Every stage always runs; each one receives the request enriched by the previous one and returns a further-enriched copy.

## Decorator

**Where:** `routing.HandlerFunc` → `routing.MethodHandler` → `routing.PathHandler`

All three implement `HttpHandler`. Each outer layer wraps the inner one and adds exactly one responsibility: `MethodHandler` guards the HTTP method; `PathHandler` matches the URL path and extracts path parameters. When `r.GET(path, f)` is called, the stack is assembled as `PathHandler(MethodHandler(HandlerFunc))` — a textbook Decorator chain.

The same pattern applies to **middleware**: each `MiddlewareFunc` wraps the next `HandlerFunc`, adding pre/post logic without touching the underlying handler.

`Group` is also a Decorator of sorts — it wraps `Router` and transparently prepends a prefix and a set of middlewares to every route registered through it, while the Router itself remains unaware of groups.

## Template Method (Hook)

**Where:** `plugin.Installer`

`Router.AddPlugin` defines a fixed startup skeleton: store the plugin, then call `Install` if the plugin opts in. The framework owns the sequence; each plugin fills in its own `Install` step — or skips it entirely by not implementing the interface.

## Command (Dispatch Table)

**Where:** `CodecRegistry`

A hash map keyed by media type string stores `Codec` objects. When a request arrives, the registry looks up the key and dispatches to the matching codec — the caller never knows which implementation runs. This is the classic command dispatch table: **key → command → execute**.

## Go Implicit Interfaces — Capability Discovery Without Coupling

**Where:** Plugin system throughout

Rather than requiring plugins to declare `implements Installer` or `implements ContextInjector`, the router uses runtime type assertions (`if installer, ok := p.(plugin.Installer); ok`) to discover capabilities. This means:

- A plugin has zero knowledge of the interfaces it satisfies — it just needs matching method signatures.
- Adding a new lifecycle (a new interface) requires no changes to existing plugins.
- In contrast, OOP languages (Java, C#) would require explicit interface declarations, coupling the plugin to the framework at compile time. Go's structural typing achieves the same extensibility without the coupling.

