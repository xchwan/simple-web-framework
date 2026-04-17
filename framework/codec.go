package framework

import (
	"context"
	"mime"
	"net/http"

	"github.com/xchwan/simple-web-framework/framework/builtin"
	"github.com/xchwan/simple-web-framework/framework/plugin"
)

// codecRegistryKey 是在 context 中存取 CodecRegistry 的 key。
type codecRegistryKey struct{}

// CodecRegistry 管理 media type → Codec 的對應，並負責將自身注入 request context。
type CodecRegistry struct {
	codecs map[string]plugin.Codec
}

// NewCodecRegistry 建立一個空的 CodecRegistry。
func NewCodecRegistry() *CodecRegistry {
	return &CodecRegistry{codecs: make(map[string]plugin.Codec)}
}

// Register 新增或覆蓋一個 media type 的 Codec。
func (cr *CodecRegistry) Register(mediaType string, c plugin.Codec) {
	cr.codecs[mediaType] = c
}

// PrepareRequest 實作 RequestPreparer，將自身注入 request context。
func (cr *CodecRegistry) PrepareRequest(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), codecRegistryKey{}, cr))
}

// loadCodecRegistry 從 request context 取出 CodecRegistry。
func loadCodecRegistry(r *http.Request) *CodecRegistry {
	cr, _ := r.Context().Value(codecRegistryKey{}).(*CodecRegistry)
	return cr
}

// lookupCodec 依 Content-Type header 查找對應的 Codec，找不到時 fallback 為 JSON。
func lookupCodec(r *http.Request, contentType string) (string, plugin.Codec) {
	mt, _, _ := mime.ParseMediaType(contentType)
	if cr := loadCodecRegistry(r); cr != nil {
		if c := cr.codecs[mt]; c != nil {
			return mt, c
		}
	}
	return "application/json", &builtin.JsonCodec{}
}
