package builtin

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/xchwan/simple-web-framework/framework/plugin"
)

// JsonCodec 是 application/json 的內建 Codec。
type JsonCodec struct{}

func (c *JsonCodec) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func (c *JsonCodec) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

// TextCodec 是 text/plain 的內建 Codec。
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

var _ plugin.Codec = (*JsonCodec)(nil)
var _ plugin.Codec = (*TextCodec)(nil)
