package framework

import (
	"context"
	"net/http"
	"strings"
)

// PathHandler 只在請求的 URL Path 符合 pattern 時才將請求往下傳遞。
// pattern 支援 {param} 語法，例如 /api/users/{id}。
type PathHandler struct {
	segments []string
	wrapped  HttpHandler
}

// NewPathHandler 建立一個 PathHandler，包裝 wrapped handler。
func NewPathHandler(pattern string, wrapped HttpHandler) *PathHandler {
	return &PathHandler{
		segments: strings.Split(strings.Trim(pattern, "/"), "/"),
		wrapped:  wrapped,
	}
}

// Handle 實作 HttpHandler。路徑不符回傳 NotMatched，符合則將 path params 存入 context 後交給 wrapped。
func (d *PathHandler) Handle(w http.ResponseWriter, r *http.Request) HandleResult {
	params, ok := matchPath(d.segments, r.URL.Path)
	if !ok {
		return NotMatched
	}
	r = r.WithContext(context.WithValue(r.Context(), pathParamsKey{}, params))
	return d.wrapped.Handle(w, r)
}

// matchPath 比對 pattern segments 與實際路徑，回傳擷取到的 path params。
func matchPath(segments []string, path string) (map[string]string, bool) {
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) != len(pathSegments) {
		return nil, false
	}
	params := map[string]string{}
	for i, seg := range segments {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			params[seg[1:len(seg)-1]] = pathSegments[i]
		} else if seg != pathSegments[i] {
			return nil, false
		}
	}
	return params, true
}

// pathParamsKey 是在 context 中存取 path params 的 key。
type pathParamsKey struct{}

// PathParam 從 request context 取出指定的 path parameter。
func PathParam(r *http.Request, key string) string {
	if params, ok := r.Context().Value(pathParamsKey{}).(map[string]string); ok {
		return params[key]
	}
	return ""
}
