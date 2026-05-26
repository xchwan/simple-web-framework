# 🗜️ Codec Extension

## Enable XML Support

The framework ships with a built-in `XmlCodec`. Install it to handle `application/xml` requests and responses.

```go
import "github.com/xchwan/simple-web-framework/plugin"

router.AddPlugin(&plugin.XmlCodec{})
```

## Adding a Custom Media Type

Implement `plugin.Codec` and register it via `Installer`:

```go
import (
    "io"
    "reflect"

    "github.com/xchwan/simple-web-framework/plugin"
)

type MsgpackCodec struct{}

func (c *MsgpackCodec) Install(ctx plugin.PluginContext) {
    ctx[reflect.TypeOf((*plugin.CodecRegistry)(nil))].(*plugin.CodecRegistry).
        Register("application/msgpack", c)
}

func (c *MsgpackCodec) Encode(w io.Writer, v any) error {
    return msgpack.NewEncoder(w).Encode(v)
}

func (c *MsgpackCodec) Decode(r io.Reader, v any) error {
    return msgpack.NewDecoder(r).Decode(v)
}

// Install
router.AddPlugin(&MsgpackCodec{})
```

## How It Works

`CodecRegistry` is a hash map keyed by media type string. At request time:

- `ParseRequest` / `ParseOrRespond` look up the `Content-Type` header → select decoder
- `Respond` looks up the `Accept` header → select encoder

This is the **Command dispatch table** pattern: key → codec → execute.
