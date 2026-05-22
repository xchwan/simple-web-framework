// Package scope provides lifecycle management strategies for the IoC container.
// To add a custom lifecycle, implement the Scope interface in a new file.
package scope

import "context"

// Scope defines the lifecycle strategy for a registered dependency.
type Scope interface {
	Resolve(ctx context.Context, factory func() any) any
}
