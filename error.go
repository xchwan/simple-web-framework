package framework

import (
	"context"
	"errors"
	"net/http"

	"github.com/xchwan/simple-web-framework/plugin"
	"github.com/xchwan/simple-web-framework/routing"
)

// ErrBadRequest is a framework-level sentinel representing a malformed request (e.g. JSON parse failure).
// HandleError maps it to 400 automatically — no ExceptionMapperPlugin configuration required.
var ErrBadRequest = errors.New("bad request")

// ErrorHandlerFunc handles routing-layer HTTP errors (404, 405).
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, statusCode int)

// ErrorBody is the unified JSON error response structure.
type ErrorBody struct {
	Message string `json:"message"`
}

// Error constructs an ErrorBody with the given message.
func Error(message string) ErrorBody {
	return ErrorBody{Message: message}
}

// ===== errorHandler context =====

type errorHandlerKey struct{}

func storeErrorHandler(r *http.Request, f ErrorHandlerFunc) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), errorHandlerKey{}, f))
}

func loadErrorHandler(r *http.Request) ErrorHandlerFunc {
	f, _ := r.Context().Value(errorHandlerKey{}).(ErrorHandlerFunc)
	return f
}

// handleRoutingError calls the error handler with 404 or 405 based on the best routing match.
func handleRoutingError(w http.ResponseWriter, r *http.Request, f ErrorHandlerFunc, best routing.HandleResult) {
	switch best {
	case routing.PathMatched:
		f(w, r, http.StatusMethodNotAllowed)
	default:
		f(w, r, http.StatusNotFound)
	}
}

// HandleError converts a Go error into an HTTP response, trying in order:
//  1. ExceptionMapperPlugin custom rules
//  2. Framework defaults (ErrBadRequest → 400)
//  3. Fallback 500
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if mapper := plugin.LoadExceptionMapper(r); mapper != nil {
		if code, msg, ok := mapper.Map(err); ok {
			Respond(w, r, code, Error(msg))
			return
		}
	}
	if errors.Is(err, ErrBadRequest) {
		Respond(w, r, http.StatusBadRequest, Error("bad request"))
		return
	}
	Respond(w, r, http.StatusInternalServerError, Error(err.Error()))
}
