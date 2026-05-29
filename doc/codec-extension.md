# 🗜️ Codec Extension

## Built-in Optional Codecs

The following codecs ship with the framework but must be installed manually:

```go
import "github.com/xchwan/simple-web-framework/plugin/codec"

router.AddPlugin(&codec.XmlCodec{})      // application/xml
router.AddPlugin(&codec.YamlCodec{})     // application/yaml
router.AddPlugin(&codec.MsgpackCodec{})  // application/msgpack
```

| Codec | Media Type | Default |
|-------|-----------|---------|
| `JsonCodec` | `application/json` | ✅ Auto |
| `TextCodec` | `text/plain` | ✅ Auto |
| `XmlCodec` | `application/xml` | 🔧 Manual |
| `YamlCodec` | `application/yaml` | 🔧 Manual |
| `MsgpackCodec` | `application/msgpack` | 🔧 Manual |

## Adding a Custom Media Type

Implement `codec.Codec` and register it via `plugin.Installer`:

```go
import (
    "io"
    "reflect"

    "github.com/xchwan/simple-web-framework/plugin"
    "github.com/xchwan/simple-web-framework/plugin/codec"
)

type TomlCodec struct{}

func (c *TomlCodec) Install(ctx plugin.PluginContext) {
    ctx[reflect.TypeOf((*codec.CodecRegistry)(nil))].(*codec.CodecRegistry).
        Register("application/toml", c)
}

func (c *TomlCodec) Encode(w io.Writer, v any) error { ... }
func (c *TomlCodec) Decode(r io.Reader, v any) error { ... }

router.AddPlugin(&TomlCodec{})
```

## How It Works

`CodecRegistry` is a hash map keyed by media type string. At request time:

- `ParseRequest` / `ParseOrRespond` look up the `Content-Type` header → select decoder
- `Respond` looks up the `Content-Type` header → select encoder

Falls back to JSON when no matching codec is found.

This is the **Command dispatch table** pattern: key → codec → execute.
