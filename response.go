package framework

import (
	"net/http"

	"github.com/xchwan/simple-web-framework/hook"
	"github.com/xchwan/simple-web-framework/plugin"
)

// Respond selects a Codec based on the Content-Type header, serializes body, and writes the response.
// For 204 No Content the Content-Type header is omitted and no body is written.
// Routing-layer 404/405 errors bypass this function and go directly through the error handler.
func Respond(w http.ResponseWriter, r *http.Request, statusCode int, body any) {
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return
	}
	mt, c := plugin.Lookup(r, r.Header.Get("Content-Type"))
	w.Header().Set("Content-Type", mt)
	w.WriteHeader(statusCode)
	if body != nil {
		c.Encode(w, body)
	}
	hook.Load(r).NotifyRespond(r, statusCode)
}
