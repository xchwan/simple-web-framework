package codec

import (
	"io"
	"reflect"

	"github.com/xchwan/simple-web-framework/plugin"
	"gopkg.in/yaml.v3"
)

// YamlCodec provides application/yaml serialization and deserialization.
// It implements plugin.Installer to register itself into the CodecRegistry at startup.
//
//	router.AddPlugin(&codec.YamlCodec{})
type YamlCodec struct{}

func (c *YamlCodec) Install(ctx plugin.PluginContext) {
	ctx[reflect.TypeOf((*CodecRegistry)(nil))].(*CodecRegistry).Register("application/yaml", c)
}

func (c *YamlCodec) Encode(w io.Writer, v any) error {
	return yaml.NewEncoder(w).Encode(v)
}

func (c *YamlCodec) Decode(r io.Reader, v any) error {
	return yaml.NewDecoder(r).Decode(v)
}
