package framework

import (
	"net/http"

	"github.com/xchwan/simple-web-framework/plugin/codec"
)

// ParseRequest selects a Codec based on the Content-Type header and decodes the request body into v.
// On failure the raw error is returned and the caller is responsible for responding.
func ParseRequest(r *http.Request, v any) error {
	_, c := codec.Lookup(r, r.Header.Get("Content-Type"))
	return c.Decode(r.Body, v)
}

// ParseOrRespond behaves like ParseRequest but automatically calls HandleError on decode failure.
// HandleError tries ExceptionMapperPlugin rules → ErrBadRequest default (400) → fallback 500.
// The caller only needs to check whether the returned error is nil and return early if not.
func ParseOrRespond(w http.ResponseWriter, r *http.Request, v any) error {
	_, c := codec.Lookup(r, r.Header.Get("Content-Type"))
	if err := c.Decode(r.Body, v); err != nil {
		HandleError(w, r, ErrBadRequest)
		return ErrBadRequest
	}
	return nil
}
