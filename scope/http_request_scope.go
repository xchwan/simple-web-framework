package scope

import (
	"context"
	"net/http"
	"sync"
)

// StoreKey 是在 context 中存取 RequestScopeStore 的 key。
type StoreKey struct{}

// RequestScopeStore 儲存同一個 HTTP request 內各個 HttpRequestScope 的實體。
type RequestScopeStore struct {
	mu        sync.Mutex
	instances map[any]any
}

// HttpRequestScope 在每一次 HTTP request 期間只創建一個實體。
// 不同 request 各自擁有獨立的實體。
type HttpRequestScope struct{}

// NewHttpRequestScope 建立一個 HttpRequestScope。
func NewHttpRequestScope() *HttpRequestScope {
	return &HttpRequestScope{}
}

func (s *HttpRequestScope) Resolve(ctx context.Context, factory func() any) any {
	store, ok := ctx.Value(StoreKey{}).(*RequestScopeStore)
	if !ok {
		return factory()
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

// InjectRequestScopeStore 在每個 HTTP request 開始時將空的 store 注入 context。
// 由 Router.ServeHTTP 呼叫。
func InjectRequestScopeStore(r *http.Request) *http.Request {
	store := &RequestScopeStore{instances: make(map[any]any)}
	return r.WithContext(context.WithValue(r.Context(), StoreKey{}, store))
}
