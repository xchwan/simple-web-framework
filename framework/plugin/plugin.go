package plugin

import (
	"io"
	"net/http"
)

// Codec 負責特定 media type 的序列化與反序列化。
type Codec interface {
	Encode(w io.Writer, v any) error
	Decode(r io.Reader, v any) error
}

// Registrar 是插件安裝時使用的註冊介面，避免直接依賴 Router。
type Registrar interface {
	RegisterCodec(mediaType string, c Codec)
}

// Plugin 是框架插件的統一介面。
type Plugin interface {
	Install(r Registrar)
}

// RequestPreparer 由需要在每個 request 注入 context 的插件實作。
type RequestPreparer interface {
	PrepareRequest(r *http.Request) *http.Request
}

