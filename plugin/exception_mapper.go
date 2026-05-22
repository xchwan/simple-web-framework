package plugin

import (
	"context"
	"errors"
	"net/http"
)

type exceptionMapperKey struct{}

// exceptionRule stores the HTTP status code and message for a single error mapping.
type exceptionRule struct {
	statusCode int
	message    string
}

// ExceptionMapperPlugin maps errors to HTTP status codes and response messages using a hash table.
type ExceptionMapperPlugin struct {
	rules map[error]exceptionRule
}

// NewExceptionMapperPlugin creates an empty ExceptionMapperPlugin.
func NewExceptionMapperPlugin() *ExceptionMapperPlugin {
	return &ExceptionMapperPlugin{rules: make(map[error]exceptionRule)}
}

// On registers an error → statusCode + message mapping. Supports method chaining.
func (p *ExceptionMapperPlugin) On(err error, statusCode int, message string) *ExceptionMapperPlugin {
	p.rules[err] = exceptionRule{statusCode, message}
	return p
}

// Map looks up the status code and message for the given error.
// Tries direct map lookup first (O(1)), then falls back to errors.Is for wrapped errors (O(n)).
func (p *ExceptionMapperPlugin) Map(err error) (statusCode int, message string, ok bool) {
	if rule, found := p.rules[err]; found {
		return rule.statusCode, rule.message, true
	}
	for target, rule := range p.rules {
		if errors.Is(err, target) {
			return rule.statusCode, rule.message, true
		}
	}
	return 0, "", false
}

// Inject implements ContextInjector, storing the mapper in the request context.
func (p *ExceptionMapperPlugin) Inject(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), exceptionMapperKey{}, p))
}

// LoadExceptionMapper retrieves the ExceptionMapperPlugin from the request context.
func LoadExceptionMapper(r *http.Request) *ExceptionMapperPlugin {
	m, _ := r.Context().Value(exceptionMapperKey{}).(*ExceptionMapperPlugin)
	return m
}
