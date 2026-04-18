package plugin

import (
	"net/http"
	"reflect"
)

// PluginContext 是插件安裝時可存取的所有已註冊資源，以型別為 key。
type PluginContext map[reflect.Type]any

// Installer 由需要在安裝時做初始化（如向 CodecRegistry 註冊）的插件實作。
type Installer interface {
	Install(ctx PluginContext)
}

// ContextInjector 由需要在每個 request 注入 context 的插件實作。
type ContextInjector interface {
	Inject(r *http.Request) *http.Request
}
