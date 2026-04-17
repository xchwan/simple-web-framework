package framework

import (
	"context"
	"mime"
	"net/http"

	"github.com/xchwan/simple-web-framework/framework/builtin"
	"github.com/xchwan/simple-web-framework/framework/plugin"
)

// codecKey 是在 context 中存取 codec registry 的 key。
type codecKey struct{}

// injectCodecs 將 codec registry 注入 request context。
func injectCodecs(r *http.Request, codecs map[string]plugin.Codec) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), codecKey{}, codecs))
}

// lookupCodec 依 Content-Type header 查找對應的 Codec，找不到時 fallback 為 JSON。
func lookupCodec(r *http.Request, contentType string) (string, plugin.Codec) {
	mt, _, _ := mime.ParseMediaType(contentType)
	if codecs, ok := r.Context().Value(codecKey{}).(map[string]plugin.Codec); ok {
		if c, ok := codecs[mt]; ok {
			return mt, c
		}
	}
	return "application/json", &builtin.JsonCodec{}
}
