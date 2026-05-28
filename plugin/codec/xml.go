package codec

import (
	"encoding/xml"
	"io"
	"reflect"

	"github.com/xchwan/simple-web-framework/plugin"
)

// XmlCodec provides application/xml serialization and deserialization.
// It implements plugin.Installer to register itself into the CodecRegistry at startup.
//
//	router.AddPlugin(&codec.XmlCodec{})
type XmlCodec struct{}

func (c *XmlCodec) Install(ctx plugin.PluginContext) {
	ctx[reflect.TypeOf((*CodecRegistry)(nil))].(*CodecRegistry).Register("application/xml", c)
}

func (c *XmlCodec) Encode(w io.Writer, v any) error {
	return xml.NewEncoder(w).Encode(v)
}

func (c *XmlCodec) Decode(r io.Reader, v any) error {
	return xml.NewDecoder(r).Decode(v)
}
