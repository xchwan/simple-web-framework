package scope

import "context"

// PrototypeScope creates a new instance on every Resolve call.
type PrototypeScope struct{}

// NewPrototypeScope creates a PrototypeScope.
func NewPrototypeScope() *PrototypeScope {
	return &PrototypeScope{}
}

func (s *PrototypeScope) Resolve(_ context.Context, factory func() any) any {
	return factory()
}
