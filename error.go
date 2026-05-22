package framework

import (
	"context"
	"errors"
	"net/http"

	"github.com/xchwan/simple-web-framework/plugin"
	"github.com/xchwan/simple-web-framework/routing"
)

// ErrBadRequest 是 framework 層的預設 sentinel，代表 request 格式錯誤（如 JSON 解析失敗）。
// HandleError 會自動將它對應到 400，不需要在 ExceptionMapperPlugin 額外設定。
var ErrBadRequest = errors.New("bad request")

// ErrorHandlerFunc 統一處理所有 routing 層的 HTTP 錯誤回應（404、405）。
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, statusCode int)

// ErrorBody 是統一的 JSON 錯誤回應格式。
type ErrorBody struct {
	Message string `json:"message"`
}

// Error 建立一個帶有訊息的 ErrorBody。
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

// handleRoutingError 依路由匹配結果呼叫 errorHandler，回傳 404 或 405。
func handleRoutingError(w http.ResponseWriter, r *http.Request, f ErrorHandlerFunc, best routing.HandleResult) {
	switch best {
	case routing.PathMatched:
		f(w, r, http.StatusMethodNotAllowed)
	default:
		f(w, r, http.StatusNotFound)
	}
}

// HandleError 將 error 轉成對應的 HTTP 回應，依序嘗試：
//  1. ExceptionMapperPlugin 的自訂規則
//  2. Framework 預設 mapping（如 ErrBadRequest → 400）
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
