package plugin

import (
	"context"
	"errors"
	"net/http"
)

type exceptionMapperKey struct{}

// exceptionRule 儲存一條 error → statusCode + message 的對應規則。
type exceptionRule struct {
	statusCode int
	message    string
}

// ExceptionMapperPlugin 以 hash table 將 error 對應到 HTTP status code 與回應訊息。
type ExceptionMapperPlugin struct {
	rules map[error]exceptionRule
}

// NewExceptionMapperPlugin 建立一個空的 ExceptionMapperPlugin。
func NewExceptionMapperPlugin() *ExceptionMapperPlugin {
	return &ExceptionMapperPlugin{rules: make(map[error]exceptionRule)}
}

// On 新增一條 error → statusCode + message 的對應規則。
func (p *ExceptionMapperPlugin) On(err error, statusCode int, message string) *ExceptionMapperPlugin {
	p.rules[err] = exceptionRule{statusCode, message}
	return p
}

// Map 查找 error 對應的 status code 與訊息。
// 先直接查 map（O(1)），找不到再用 errors.Is 處理 wrapped error（O(n)）。
func (p *ExceptionMapperPlugin) Map(err error) (statusCode int, message string, ok bool) {
	if rule, ok := p.rules[err]; ok {
		return rule.statusCode, rule.message, true
	}
	for target, rule := range p.rules {
		if errors.Is(err, target) {
			return rule.statusCode, rule.message, true
		}
	}
	return 0, "", false
}

// Install 實作 Plugin 介面，ExceptionMapperPlugin 透過 PrepareRequest 自行注入，無需向 Registrar 登記。
func (p *ExceptionMapperPlugin) Install(_ Registrar) {}

// PrepareRequest 實作 RequestPreparer，將自身注入 request context。
func (p *ExceptionMapperPlugin) PrepareRequest(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), exceptionMapperKey{}, p))
}

// LoadExceptionMapper 從 request context 取出 ExceptionMapperPlugin。
func LoadExceptionMapper(r *http.Request) *ExceptionMapperPlugin {
	m, _ := r.Context().Value(exceptionMapperKey{}).(*ExceptionMapperPlugin)
	return m
}
