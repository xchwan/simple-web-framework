package plugin

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
)

type codecRegistryKey struct{}

// CodecRegistry 管理 media type → Codec 的對應，並負責將自身注入 request context。
type CodecRegistry struct {
	codecs map[string]Codec
}

// NewCodecRegistry 建立一個空的 CodecRegistry。
func NewCodecRegistry() *CodecRegistry {
	return &CodecRegistry{codecs: make(map[string]Codec)}
}

// Register 新增或覆蓋一個 media type 的 Codec。
func (cr *CodecRegistry) Register(mediaType string, c Codec) {
	cr.codecs[mediaType] = c
}

// Inject 實作 ContextInjector，將自身注入 request context。
func (cr *CodecRegistry) Inject(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), codecRegistryKey{}, cr))
}

func loadCodecRegistry(r *http.Request) *CodecRegistry {
	cr, _ := r.Context().Value(codecRegistryKey{}).(*CodecRegistry)
	return cr
}

// Lookup 依 Content-Type header 查找對應的 Codec，找不到時 fallback 為 JSON。
func Lookup(r *http.Request, contentType string) (string, Codec) {
	mt, _, _ := mime.ParseMediaType(contentType)
	if cr := loadCodecRegistry(r); cr != nil {
		if c := cr.codecs[mt]; c != nil {
			return mt, c
		}
	}
	return "application/json", &jsonFallback{}
}

// Codec 負責特定 media type 的序列化與反序列化。
type Codec interface {
	Encode(w io.Writer, v any) error
	Decode(r io.Reader, v any) error
}

// jsonFallback 是 Lookup 找不到匹配 Codec 時的預設實作。
type jsonFallback struct{}

func (c *jsonFallback) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func (c *jsonFallback) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}
