package scope

import (
	"context"
	"sync"
)

// SingletonScope 在整個程式執行期間只創建一個實體。
type SingletonScope struct {
	once     sync.Once
	instance any
}

// NewSingletonScope 建立一個 SingletonScope。
func NewSingletonScope() *SingletonScope {
	return &SingletonScope{}
}

func (s *SingletonScope) Resolve(_ context.Context, factory func() any) any {
	s.once.Do(func() {
		s.instance = factory()
	})
	return s.instance
}
