// Package scope 提供 IoC Container 的生命週期管理策略。
// 新增自訂生命週期只需實作 Scope 介面，新增一個檔案即可。
package scope

import "context"

// Scope 定義依賴的生命週期管理策略。
type Scope interface {
	Resolve(ctx context.Context, factory func() any) any
}
