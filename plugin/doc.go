package plugin

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

// NoBody is used as a type parameter when a route has no request or response body.
//
//	router.GET("/api/users", h.List, plugin.Doc[plugin.NoBody, []UserResponse](docs))
type NoBody = struct{}

// Doc stores request/response type metadata in docs keyed by the handler's function pointer,
// then returns the original handler unchanged. When the router registers the route,
// it calls docs.OnRegister which matches the pointer, completes the record with method and path,
// and clears the pending entry.
//
//	docs := plugin.NewDocPlugin()
//	router.AddPlugin(docs)
//	router.POST("/api/users", h.Create, plugin.Doc[CreateUserRequest, UserResponse](docs))
func Doc[Req, Resp any](docs *DocPlugin, f HandlerFunc) HandlerFunc {
	var req Req
	var resp Resp
	ptr := reflect.ValueOf(f).Pointer()
	docs.pending.Store(ptr, docMeta{
		requestType:  reflect.TypeOf(req),
		responseType: reflect.TypeOf(resp),
	})
	return f
}

type docMeta struct {
	requestType  reflect.Type
	responseType reflect.Type
}

// DocPlugin collects route metadata at registration time and serves interactive
// API documentation powered by Swagger UI.
//
//	docs := plugin.NewDocPlugin()
//	router.AddPlugin(docs)
//	router.POST("/api/users", h.Create, plugin.Doc[CreateUserRequest, UserResponse](docs))
//	router.GET("/docs",         docs.UIHandler())
//	router.GET("/openapi.json", docs.SpecHandler())
type DocPlugin struct {
	pending sync.Map // map[uintptr]docMeta
	mu      sync.RWMutex
	routes  []routeDoc
}

// NewDocPlugin creates a new, empty DocPlugin.
func NewDocPlugin() *DocPlugin {
	return &DocPlugin{}
}

// OnRegister implements RouteHook. It is called by the router for every registered route.
// If the handler was wrapped with Doc[Req, Resp], the pending metadata is matched by
// function pointer, recorded with the given method and path, then removed from pending.
func (d *DocPlugin) OnRegister(method, path string, f HandlerFunc) {
	ptr := reflect.ValueOf(f).Pointer()
	val, ok := d.pending.LoadAndDelete(ptr)
	if !ok {
		return
	}
	meta := val.(docMeta)
	doc := routeDoc{Method: method, Path: path}
	if meta.requestType != nil && meta.requestType != noBodyType {
		doc.Request = typeSchema(meta.requestType)
	}
	if meta.responseType != nil && meta.responseType != noBodyType {
		doc.Response = typeSchema(meta.responseType)
	}
	d.mu.Lock()
	d.routes = append(d.routes, doc)
	d.mu.Unlock()
}

// SpecHandler returns a HandlerFunc that serves the OpenAPI 3.0 JSON spec.
//
//	router.GET("/openapi.json", docs.SpecHandler())
func (d *DocPlugin) SpecHandler() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d.buildSpec())
	}
}

// UIHandler returns a HandlerFunc that serves the Swagger UI page.
//
//	router.GET("/docs", docs.UIHandler())
func (d *DocPlugin) UIHandler() HandlerFunc {
	html := `<!DOCTYPE html>
<html>
<head>
  <title>API Docs</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/openapi.json",
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout"
    })
  </script>
</body>
</html>`
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	}
}

// --- internal ---

var noBodyType = reflect.TypeOf(struct{}{})

type routeDoc struct {
	Method   string         `json:"method"`
	Path     string         `json:"path"`
	Request  map[string]any `json:"request,omitempty"`
	Response map[string]any `json:"response,omitempty"`
}

func (d *DocPlugin) buildSpec() map[string]any {
	d.mu.RLock()
	defer d.mu.RUnlock()

	paths := make(map[string]any)
	for _, r := range d.routes {
		if _, ok := paths[r.Path]; !ok {
			paths[r.Path] = make(map[string]any)
		}
		pathItem := paths[r.Path].(map[string]any)
		pathItem[strings.ToLower(r.Method)] = buildOperation(r)
	}
	return map[string]any{
		"openapi": "3.0.0",
		"info":    map[string]any{"title": "API", "version": "1.0.0"},
		"paths":   paths,
	}
}

func buildOperation(r routeDoc) map[string]any {
	op := make(map[string]any)
	if r.Request != nil {
		op["requestBody"] = map[string]any{
			"required": true,
			"content":  map[string]any{"application/json": map[string]any{"schema": r.Request}},
		}
	}
	resp := map[string]any{"description": "Success"}
	if r.Response != nil {
		resp["content"] = map[string]any{"application/json": map[string]any{"schema": r.Response}}
	}
	op["responses"] = map[string]any{"200": resp}
	return op
}

func typeSchema(t reflect.Type) map[string]any {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		props := make(map[string]any)
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				name = f.Name
			}
			props[name] = map[string]any{"type": kindToJSONType(f.Type)}
		}
		if len(props) == 0 {
			return nil
		}
		return map[string]any{"type": "object", "properties": props}
	case reflect.Slice:
		return map[string]any{"type": "array", "items": typeSchema(t.Elem())}
	default:
		return map[string]any{"type": kindToJSONType(t)}
	}
}

func kindToJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	default:
		return "object"
	}
}
