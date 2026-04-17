package plugin

import (
	"encoding/xml"
	"io"
)

// XmlMediaTypePlugin 提供 application/xml 的序列化/反序列化支援。
type XmlMediaTypePlugin struct{}

func (p *XmlMediaTypePlugin) Install(r Registrar) {
	r.RegisterCodec("application/xml", &xmlCodec{})
}

type xmlCodec struct{}

func (c *xmlCodec) Encode(w io.Writer, v any) error {
	return xml.NewEncoder(w).Encode(v)
}

func (c *xmlCodec) Decode(r io.Reader, v any) error {
	return xml.NewDecoder(r).Decode(v)
}
