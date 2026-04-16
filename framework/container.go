package framework

import (
	"context"
	"net/http"
	"sync"

	"github.com/xchwan/simple-web-framework/framework/scope"
)

// containerKey 是在 context 中存取 Container 的 key。
type containerKey struct{}

// registration 儲存一個依賴的工廠函式與生命週期。
type registration struct {
	factory func() any
	scope   scope.Scope
}

// Container 管理依賴的工廠函式與生命週期。
type Container struct {
	mu            sync.RWMutex
	registrations map[string]*registration
}

// NewContainer 建立一個空的 Container。
func NewContainer() *Container {
	return &Container{registrations: make(map[string]*registration)}
}

// Register 向容器註冊一個依賴。
// s 省略時預設使用 SingletonScope。
func (c *Container) Register(name string, factory func() any, s ...scope.Scope) {
	c.mu.Lock()
	defer c.mu.Unlock()
	sc := scope.Scope(scope.NewSingletonScope())
	if len(s) > 0 {
		sc = s[0]
	}
	c.registrations[name] = &registration{factory: factory, scope: sc}
}

func (c *Container) get(ctx context.Context, name string) any {
	c.mu.RLock()
	r, ok := c.registrations[name]
	c.mu.RUnlock()
	if !ok {
		return nil
	}
	return r.scope.Resolve(ctx, r.factory)
}

// injectContainer 將 container 注入 request context。
func injectContainer(r *http.Request, c *Container) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), containerKey{}, c))
}

// Get 從 request context 中的 Container 取出指定名稱的依賴實體。
func Get[T any](r *http.Request, name string) T {
	c := r.Context().Value(containerKey{}).(*Container)
	return c.get(r.Context(), name).(T)
}
