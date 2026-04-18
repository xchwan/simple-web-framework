package framework

import (
	"net/http"

	"github.com/xchwan/simple-web-framework/framework/plugin"
)

// ParseRequest 依 Content-Type header 選擇對應的 Codec，將 request body 解析到 v。
func ParseRequest(r *http.Request, v any) error {
	_, c := plugin.Lookup(r, r.Header.Get("Content-Type"))
	return c.Decode(r.Body, v)
}
