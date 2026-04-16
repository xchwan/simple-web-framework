// Package framework 提供輕量的 HTTP 路由框架核心元件。
package framework

import "net/http"

// HandleResult 表示 handler 對一個請求的處理結果。
type HandleResult int

const (
	// NotMatched 表示此 handler 完全不匹配（路徑不符），Router 繼續往下嘗試。
	NotMatched HandleResult = iota
	// PathMatched 表示路徑符合但 Method 不符，Router 應回傳 405。
	PathMatched HandleResult = iota
	// Handled 表示已完整處理請求，Router 應停止往下嘗試。
	Handled HandleResult = iota
)

// HttpHandler 是框架內所有元件共同實作的介面。
type HttpHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) HandleResult
}

// HandlerFunc 是函式型別的轉接器，讓普通函式滿足 HttpHandler 介面。
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Handle 呼叫函式本身，並回傳 Handled。
func (f HandlerFunc) Handle(w http.ResponseWriter, r *http.Request) HandleResult {
	f(w, r)
	return Handled
}
