package scope

import (
	"context"
	"net/http"
	"sync"
)

// StoreKey is the context key used to store and retrieve the RequestScopeStore.
type StoreKey struct{}

// RequestScopeStore caches instances created by HttpRequestScope for a single HTTP request.
type RequestScopeStore struct {
	mu        sync.Mutex
	instances map[any]any
}

// HttpRequestScope creates one instance per HTTP request.
// Distinct requests each get their own independent instance.
//
// The unexported _ byte field makes this a non-zero-size struct.
// Without it, Go would point all instances to runtime.zerobase,
// making every *HttpRequestScope identical as a map key and causing
// different bindings to collide in the RequestScopeStore.
type HttpRequestScope struct {
	_ byte
}

// NewHttpRequestScope creates an HttpRequestScope.
// Each call returns a pointer with a unique address, used as the store key.
func NewHttpRequestScope() *HttpRequestScope {
	return &HttpRequestScope{}
}

func (s *HttpRequestScope) Resolve(ctx context.Context, factory func() any) any {
	store, ok := ctx.Value(StoreKey{}).(*RequestScopeStore)
	if !ok {
		panic("HttpRequestScope.Resolve called without a request context — this is a framework bug")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if instance, exists := store.instances[s]; exists {
		return instance
	}
	instance := factory()
	store.instances[s] = instance
	return instance
}

// InjectRequestScopeStore attaches a fresh empty store to the request context at the start of each request.
// Called by Router.injectContext.
func InjectRequestScopeStore(r *http.Request) *http.Request {
	store := &RequestScopeStore{instances: make(map[any]any)}
	return r.WithContext(context.WithValue(r.Context(), StoreKey{}, store))
}
