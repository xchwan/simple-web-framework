package framework

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/xchwan/simple-web-framework/scope"
)

// containerKey is the context key used to store and retrieve the Container.
type containerKey struct{}

// registration holds the factory function and lifecycle scope for a single dependency.
type registration struct {
	factory func() any
	scope   scope.Scope
}

// Container manages dependency registrations and their lifecycles.
type Container struct {
	mu            sync.RWMutex
	registrations map[string]registration
}

// NewContainer creates an empty Container.
func NewContainer() *Container {
	return &Container{registrations: make(map[string]registration)}
}

// Register adds a dependency to the container.
// Defaults to SingletonScope when no scope is provided.
func (c *Container) Register(name string, factory func() any, s ...scope.Scope) {
	c.mu.Lock()
	defer c.mu.Unlock()
	sc := scope.Scope(scope.NewSingletonScope())
	if len(s) > 0 {
		sc = s[0]
	}
	c.registrations[name] = registration{factory: factory, scope: sc}
}

// Resolve retrieves the named dependency, delegating creation to the registered scope.
func (c *Container) Resolve(ctx context.Context, name string) any {
	c.mu.RLock()
	r, ok := c.registrations[name]
	c.mu.RUnlock()
	if !ok {
		panic(fmt.Sprintf(
			"dependency %q not found — register it with router.Bind(%q, func() any { return ... })",
			name, name,
		))
	}
	return r.scope.Resolve(ctx, r.factory)
}

func (c *Container) get(ctx context.Context, name string) any {
	return c.Resolve(ctx, name)
}

// injectContainer stores the container in the request context and initialises the HttpRequestScope store.
func injectContainer(r *http.Request, c *Container) *http.Request {
	r = scope.InjectRequestScopeStore(r)
	return r.WithContext(context.WithValue(r.Context(), containerKey{}, c))
}

// Get retrieves a named dependency from the container stored in the request context.
func Get[T any](r *http.Request, name string) T {
	c := r.Context().Value(containerKey{}).(*Container)
	return c.get(r.Context(), name).(T)
}
