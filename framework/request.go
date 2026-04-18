package framework

import (
	"net/http"

	"github.com/xchwan/simple-web-framework/framework/plugin"
)

// ParseRequest 依 Content-Type header 選擇對應的 Codec，將 request body 解析到 v。
// 解析失敗時回傳原始 error，由呼叫方自行處理。
func ParseRequest(r *http.Request, v any) error {
	_, c := plugin.Lookup(r, r.Header.Get("Content-Type"))
	return c.Decode(r.Body, v)
}

// ParseOrRespond 與 ParseRequest 相同，但解析失敗時會自動呼叫 HandleError 寫入回應。
// HandleError 依序嘗試 ExceptionMapperPlugin → framework 預設（ErrBadRequest → 400）→ 500。
// 呼叫方只需判斷回傳的 error 是否為 nil 來決定是否 return，不需要自行處理錯誤。
func ParseOrRespond(w http.ResponseWriter, r *http.Request, v any) error {
	_, c := plugin.Lookup(r, r.Header.Get("Content-Type"))
	if err := c.Decode(r.Body, v); err != nil {
		HandleError(w, r, ErrBadRequest)
		return ErrBadRequest
	}
	return nil
}
