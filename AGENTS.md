# AGENTS.md -- go-vnc

## Setup
- `go test ./...` from repo root -- all tests use in-memory `MockConn`, no external services
- CI runs `staticcheck ./...` then `go vet ./...` then `go test ./...` (push/PR to master)
- go.mod: `github.com/kward/go-vnc`, go 1.24

## Code generation
- `go generate ./...` -- stringer generates `*_string.go` enum files (already committed)
- Install: `go install golang.org/x/tools/cmd/stringer@latest`, ensure `[go env GOPATH]/bin` on PATH

## Architecture (RFC 6143 -- files mirror spec sections)
- `vncclient.go` -- entrypoint: `Connect(ctx, net.Conn, *ClientConfig) (*ClientConn, error)` runs: version -> security -> auth -> init -> encodings/pixel-format
- `handshake.go` -- protocol version + security negotiation (§7.1)
- `security.go` -- auth strategies (VNC, None)
- `initialization.go` -- client/server init (§7.3)
- `client.go` -- client->server messages: SetPixelFormat, SetEncodings, FramebufferUpdateRequest, KeyEvent, PointerEvent, ClientCutText (§7.5)
- `server.go` -- server->client messages + `ServerMessage` interface; `ListenAndHandle()` dispatches by Type() via `ClientConfig.ServerMessages` map (§7.6)
- `encodings.go` -- `Encoding` interface (Read + Marshal) + encodings/ encodings
- `encoding/wire.go` -- thin wire format marshaling/unmarshaling: `WireMarshal()` + `WireUnmarshal()` on every wire struct. Marshaling uses direct `w.WireMarshal()` (no `MarshalToWire` wrapper). Unmarshal reads from `[]byte`.
- `encoding/wire_types.go` -- wire type declarations: `MessageHeader`, `EncodingHeader` helpers + `WireTypes` registry for all VNC wire formats
- `messages/` -- wire message type constants
- `encoding/` -- wire helper types (RectDataWire, etc.), buffers
- `specs/wire.md` -- human-readable RFC 6143 wire format specification (canonical reference; every wire type must match this document)

### Wire format design principle
**All RFC-related wire format handling lives in `encoding/`.** This means:
- Every VNC wire struct has `WireMarshal() ([]byte, error)` and `WireUnmarshal(data []byte) error` methods
- Wire types for server messages (e.g., `SetColorMapEntriesWire`, `ServerCutTextWire`) include `Read(io.Reader)` methods for live deserialization
- Client structs (`PixelFormat`, etc.) delegate their `Marshal()` to the corresponding wire struct in `encoding/`
- Server `ServerMessage.Read()` implementations use the wire helpers from `encoding/` for padding, field extraction, and color entry parsing -- never manual `binary.Read`/`io.ReadFull` outside of encoding/
- No manual wire parsing in any file other than `encoding/wire.go` and `encoding/wire_types.go`
- `MarshalToWire()` and `UnmarshalFromWire()` wrappers have been removed -- always call the `WireMarshal()`/`WireUnmarshal()` method directly on the wire struct
- `marshalFields()` indirection has been removed -- inlined directly into `WireMarshal()`

## Conventions
- Wire structs: big-endian, explicit padding fields matching on-wire layout
- Wire field size comments on every struct: `// Wire: field_name(type) + field_name2(type) + ...`
- Default encodings: Raw only. For desktop resize support, include `DesktopSizePseudoEncoding` before calling SetEncodings
- Client input methods (KeyEvent, PointerEvent, ClientCutText) apply a UI settle delay; tests disable with `vnc.SetSettle(0)`
- ClientCutText: Latin-1 only, \r chars stripped per RFC 6143
- Context key `"vnc_max_proto_version"` with values `"3.3"` or `"3.8"` to cap protocol version during Connect
- Adding a server message: implement `ServerMessage` (Type, Read), add to `ClientConfig.ServerMessages` -- `ListenAndHandle` dispatches automatically
- Metrics in send/receive are approximate -- don't rely for exact byte counts
- Adding a new wire type: declare struct in `encoding/wire.go` or `encoding/wire_types.go`, implement `WireMarshal()` and `WireUnmarshal()`, update `specs/wire.md`
- Encoding structs delegate `Marshal()` to their wire equivalent in `encoding/` -- do not write raw wire logic in encoding structs
