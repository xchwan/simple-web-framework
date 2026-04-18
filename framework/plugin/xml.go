package plugin

import (
	"encoding/xml"
	"io"
	"reflect"
)

// XmlCodec 提供 application/xml 的序列化/反序列化支援。
type XmlCodec struct{}

func (c *XmlCodec) Install(ctx PluginContext) {
	ctx[reflect.TypeOf((*CodecRegistry)(nil))].(*CodecRegistry).Register("application/xml", c)
}

func (c *XmlCodec) Encode(w io.Writer, v any) error {
	return xml.NewEncoder(w).Encode(v)
}

func (c *XmlCodec) Decode(r io.Reader, v any) error {
	return xml.NewDecoder(r).Decode(v)
}
