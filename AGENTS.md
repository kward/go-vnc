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
- `encoding/` -- wire types, buffers, marshal/unmarshal
- `messages/` -- wire message type constants

## Conventions
- Wire structs: big-endian, explicit padding fields matching on-wire layout
- Default encodings: Raw only. For desktop resize support, include `DesktopSizePseudoEncoding` before calling SetEncodings
- Client input methods (KeyEvent, PointerEvent, ClientCutText) apply a UI settle delay; tests disable with `vnc.SetSettle(0)`
- ClientCutText: Latin-1 only, \r chars stripped per RFC 6143
- Context key `"vnc_max_proto_version"` with values `"3.3"` or `"3.8"` to cap protocol version during Connect
- Adding a server message: implement `ServerMessage` (Type, Read), add to `ClientConfig.ServerMessages` -- `ListenAndHandle` dispatches automatically
- Metrics in send/receive are approximate -- don't rely for exact byte counts
