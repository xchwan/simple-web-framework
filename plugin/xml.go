package plugin

import (
	"encoding/xml"
	"io"
	"reflect"
)

// XmlCodec provides application/xml serialization and deserialization.
// It implements Installer to register itself into the CodecRegistry at startup.
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
