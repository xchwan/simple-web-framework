package framework

import (
	"encoding/json"
	"net/http"
)

// ParseRequest 將 request body 的 JSON 解析到 v。
func ParseRequest(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
