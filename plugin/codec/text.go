package codec

import (
	"fmt"
	"io"
)

// TextCodec is the built-in Codec for text/plain.
// Auto-registered by the framework — no setup required.
type TextCodec struct{}

func (c *TextCodec) Encode(w io.Writer, v any) error {
	_, err := fmt.Fprint(w, v)
	return err
}

func (c *TextCodec) Decode(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if s, ok := v.(*string); ok {
		*s = string(data)
	}
	return nil
}
