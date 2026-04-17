package framework

import "net/http"

// ParseRequest 依 Content-Type header 選擇對應的 Codec，將 request body 解析到 v。
func ParseRequest(r *http.Request, v any) error {
	_, codec := lookupCodec(r, r.Header.Get("Content-Type"))
	return codec.Decode(r.Body, v)
}
