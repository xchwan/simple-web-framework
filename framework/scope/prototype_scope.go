package scope

import "context"

// PrototypeScope 每次 Resolve 都創建一個全新的實體。
type PrototypeScope struct{}

// NewPrototypeScope 建立一個 PrototypeScope。
func NewPrototypeScope() *PrototypeScope {
	return &PrototypeScope{}
}

func (s *PrototypeScope) Resolve(_ context.Context, factory func() any) any {
	return factory()
}
