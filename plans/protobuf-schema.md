# Plan: Protobuf Schema for VNC Protocol

## Goals

1. Create `proto/vnc.proto` as the single source of truth for the VNC wire format, section-for-section with RFC 6143 (§7.5 Client-to-Server, §7.6 Server-to-Client)
2. Add `go:generate` rule to produce `gen/vnc.pb.go` and `genpy/vnc_pb2.py` (future)
3. Add `encoding/wire.go` - thin VNC wire format marshaling/unmarshaling that reads/writes between the proto messages and big-endian byte streams
4. Wire it into the existing client.go / server.go without changing public API surface
5. The proto file replaces RFC cross-referencing in the code: each message's field count, types, and layout are immediately readable

## What Stays

- Public API methods (`ClientConn.SetPixelFormat`, `KeyEvent`, etc.) - unchanged
- The existing pixel format, encoding, key/button type packages - kept as-is
- Big-endian wire format via `binary.Write`/`binary.Read` - still in `encoding/wire.go`

## Structure

```
proto/
  vnc.proto                  # schema
gen/
  vnc.pb.go                  # protoc-gen-go output
encoding/
  wire.go                    # Marshal/Unmarshal between proto messages and wire bytes
go.mod                       # go run google.golang.org/protobuf/cmd/protoc-gen-go
```

## Wire Format Notes

- All message type codes stay as constants (`messages.ClientMessage`, `messages.ServerMessage`)
- `PixelFormat` stays as a local struct (too complex for direct proto mapping)
- All coord / length fields: `uint32` in proto, serialized as big-endian `uint16`/`uint32`
- Text fields: `bytes` (not `string`) to avoid UTF-8 encoding assumptions (VNC is Latin-1)
- Padding handled explicitly in `encoding/wire.go`

## Implementation Steps

1. Create `proto/vnc.proto`
2. Add protoc dependency to go.mod
3. Add `//go:generate` directive for protoc-gen-go
4. Create `encoding/wire.go` with `WireMarshal`, `WireUnmarshal` helpers
5. Wire proto messages into client.go and server.go (internal-only change)
6. Run `go generate` to produce `gen/vnc.pb.go`
7. Verify tests pass
