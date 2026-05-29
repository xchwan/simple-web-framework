package codec

import (
	"io"
	"reflect"

	"github.com/shamaton/msgpack/v2"
	"github.com/xchwan/simple-web-framework/plugin"
)

// MsgpackCodec provides application/msgpack serialization and deserialization.
// It implements plugin.Installer to register itself into the CodecRegistry at startup.
//
//	router.AddPlugin(&codec.MsgpackCodec{})
type MsgpackCodec struct{}

func (c *MsgpackCodec) Install(ctx plugin.PluginContext) {
	ctx[reflect.TypeOf((*CodecRegistry)(nil))].(*CodecRegistry).Register("application/msgpack", c)
}

func (c *MsgpackCodec) Encode(w io.Writer, v any) error {
	b, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (c *MsgpackCodec) Decode(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(b, v)
}
