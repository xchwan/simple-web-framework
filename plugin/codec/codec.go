package codec

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
)

type codecRegistryKey struct{}

// Codec handles serialization and deserialization for a specific media type.
type Codec interface {
	Encode(w io.Writer, v any) error
	Decode(r io.Reader, v any) error
}

// CodecRegistry maps media types to Codec implementations and injects itself into the request context.
// It implements plugin.ContextInjector — the framework injects it on every request automatically.
type CodecRegistry struct {
	codecs map[string]Codec
}

// NewCodecRegistry creates an empty CodecRegistry.
func NewCodecRegistry() *CodecRegistry {
	return &CodecRegistry{codecs: make(map[string]Codec)}
}

// Register adds or replaces the Codec for the given media type.
func (cr *CodecRegistry) Register(mediaType string, c Codec) {
	cr.codecs[mediaType] = c
}

// Inject implements plugin.ContextInjector, storing the registry in the request context.
func (cr *CodecRegistry) Inject(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), codecRegistryKey{}, cr))
}

func loadCodecRegistry(r *http.Request) *CodecRegistry {
	cr, _ := r.Context().Value(codecRegistryKey{}).(*CodecRegistry)
	return cr
}

// Lookup finds the Codec for the given Content-Type. Falls back to JSON when no match is found.
func Lookup(r *http.Request, contentType string) (string, Codec) {
	mt, _, _ := mime.ParseMediaType(contentType)
	if cr := loadCodecRegistry(r); cr != nil {
		if c := cr.codecs[mt]; c != nil {
			return mt, c
		}
	}
	return "application/json", &jsonFallback{}
}

// jsonFallback is the default Codec used when Lookup finds no matching media type.
type jsonFallback struct{}

func (c *jsonFallback) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func (c *jsonFallback) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}
