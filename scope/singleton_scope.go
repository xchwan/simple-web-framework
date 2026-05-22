package scope

import (
	"context"
	"sync"
)

// SingletonScope creates exactly one instance for the lifetime of the application.
type SingletonScope struct {
	once     sync.Once
	instance any
}

// NewSingletonScope creates a SingletonScope.
func NewSingletonScope() *SingletonScope {
	return &SingletonScope{}
}

func (s *SingletonScope) Resolve(_ context.Context, factory func() any) any {
	s.once.Do(func() {
		s.instance = factory()
	})
	return s.instance
}
