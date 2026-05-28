package apidoc

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/xchwan/simple-web-framework/plugin"
)

// HandlerFunc is an alias for convenience within this package.
type HandlerFunc = plugin.HandlerFunc

// NoBody is used as a type parameter when a route has no request or response body.
//
//	router.GET("/api/users", h.List, doc.Doc[doc.NoBody, []UserResponse](docs, h.List))
type NoBody = struct{}

// docMeta holds metadata collected for a single route at registration time.
// options apply themselves directly to the OpenAPI operation map, so docMeta
// never needs a new field when a new DocOption is introduced (OCP).
type docMeta struct {
	requestType  reflect.Type
	responseType reflect.Type
	options      []DocOption
}

// pending is a package-level store that maps a handler's function pointer (uintptr)
// to its docMeta. Doc writes into it; DocPlugin.RouteAdded reads from it.
// Using a package-level map means Doc does not need to receive *DocPlugin as a parameter —
// the two sides rendezvous implicitly through the function pointer key.
var pending sync.Map // map[uintptr]docMeta

// Doc annotates a handler with request/response type information for API documentation.
// It stores the metadata in a package-level map keyed by f's function pointer, then
// returns f unchanged. When the router registers the route, DocPlugin.RouteAdded is
// called with the same function pointer and picks up the metadata automatically.
//
// opts accepts a plain string (treated as Summary) or any number of DocOption values:
//
//	// plain string → summary
//	apidoc.Doc[CreateUserRequest, UserResponse](h.Create, "Register a new user")
//
//	// explicit options
//	apidoc.Doc[CreateUserRequest, UserResponse](h.Create,
//	    apidoc.Summary("Register a new user"),
//	    apidoc.Description("Email must be unique."),
//	    apidoc.Tags("users"),
//	)
func Doc[Req, Resp any](f HandlerFunc, opts ...any) HandlerFunc {
	var req Req
	var resp Resp
	meta := docMeta{
		requestType:  reflect.TypeOf(req),
		responseType: reflect.TypeOf(resp),
	}
	for _, o := range opts {
		switch v := o.(type) {
		case string:
			meta.options = append(meta.options, Summary(v))
		case DocOption:
			meta.options = append(meta.options, v)
		}
	}
	pending.Store(reflect.ValueOf(f).Pointer(), meta)
	return f
}

// routeDoc is the collected metadata for a single route, ready for serialisation.
type routeDoc struct {
	method       string
	path         string
	requestType  reflect.Type
	responseType reflect.Type
	options      []DocOption
}

// DocPlugin collects route metadata at registration time and serves interactive
// API documentation powered by Swagger UI.
//
//	import apidoc "github.com/xchwan/simple-web-framework/plugin/apidoc"
//
//	docs := apidoc.NewDocPlugin()
//	router.AddPlugin(docs)
//	router.POST("/api/users", h.Create,
//	    apidoc.Doc[CreateUserRequest, UserResponse](h.Create, "Register a new user"))
//	router.GET("/docs",         docs.UIHandler())
//	router.GET("/openapi.json", docs.SpecHandler())
type DocPlugin struct {
	mu     sync.RWMutex
	routes []routeDoc
}

// NewDocPlugin creates a new, empty DocPlugin.
func NewDocPlugin() *DocPlugin {
	return &DocPlugin{}
}

var noBodyType = reflect.TypeOf(struct{}{})

// RouteAdded implements plugin.RouteHook. Called once per route at registration time.
func (d *DocPlugin) RouteAdded(method, path string, f HandlerFunc) {
	val, ok := pending.LoadAndDelete(reflect.ValueOf(f).Pointer())
	if !ok {
		return
	}
	meta := val.(docMeta)
	d.mu.Lock()
	d.routes = append(d.routes, routeDoc{
		method:       method,
		path:         path,
		requestType:  meta.requestType,
		responseType: meta.responseType,
		options:      meta.options,
	})
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

func (d *DocPlugin) buildSpec() map[string]any {
	d.mu.RLock()
	defer d.mu.RUnlock()

	pending.Range(func(k, v any) bool {
		log.Println("[apidoc] warning: Doc[] was called but DocPlugin is not registered — call router.AddPlugin(docs)")
		return false // 只印一次
	})

	paths := make(map[string]any)
	for _, r := range d.routes {
		if _, ok := paths[r.path]; !ok {
			paths[r.path] = make(map[string]any)
		}
		paths[r.path].(map[string]any)[strings.ToLower(r.method)] = buildOperation(r)
	}
	return map[string]any{
		"openapi": "3.0.0",
		"info":    map[string]any{"title": "API", "version": "1.0.0"},
		"paths":   paths,
	}
}

func buildOperation(r routeDoc) map[string]any {
	op := make(map[string]any)

	// each option knows exactly which OpenAPI field to set
	for _, opt := range r.options {
		opt(op)
	}

	if r.requestType != nil && r.requestType != noBodyType {
		op["requestBody"] = map[string]any{
			"required": true,
			"content":  map[string]any{"application/json": map[string]any{"schema": typeSchema(r.requestType)}},
		}
	}
	resp := map[string]any{"description": "Success"}
	if r.responseType != nil && r.responseType != noBodyType {
		resp["content"] = map[string]any{"application/json": map[string]any{"schema": typeSchema(r.responseType)}}
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
