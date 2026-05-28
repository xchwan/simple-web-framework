package codec

import (
	"encoding/json"
	"io"
)

// JsonCodec is the built-in Codec for application/json.
// Auto-registered by the framework — no setup required.
type JsonCodec struct{}

func (c *JsonCodec) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func (c *JsonCodec) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}
