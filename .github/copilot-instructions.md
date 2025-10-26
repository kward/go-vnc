# Copilot instructions for go-vnc

Purpose: This repo is a VNC client library that implements RFC 6143. Files are organized to mirror the spec sections and expose a small API for negotiating a connection and sending client messages.

## Big picture architecture
- Connection flow: `Connect(ctx, net.Conn, *ClientConfig) (*ClientConn, error)` in `vncclient.go` runs, in order: protocol version (§7.1.1), security (§7.1.2/7.1.3), client init (§7.3.1), server init (§7.3.2), then sends `SetEncodings` and `SetPixelFormat`.
- Message routing: `ClientConn.ListenAndHandle()` reads a `messages.ServerMessage` byte, looks up a prototype in `ClientConfig.ServerMessages`, calls `Read(*ClientConn)` on it, then pushes the parsed message onto `ServerMessageCh`.
- Encodings: Server-to-client rectangles are represented by `Rectangle` with an `Encoding` strategy (see `encodings.go`). Raw is always supported; other encodings must be included in `ClientConn.encodings`.
- Pixel format and color: `PixelFormat` describes wire pixel layout; `Color` and `ColorMap` translate wire values. True-color vs color-mapped behavior is handled in `Color.Unmarshal`.

## Key files and how to extend
- Handshake and init: `handshake.go`, `security.go`, `initialization.go` (map to RFC §7.1–7.3).
- Client->Server messages: `client.go` (§7.5). Example: `FramebufferUpdateRequest`, `KeyEvent`, `PointerEvent`.
- Server->Client messages: `server.go` (§7.6). Includes `FramebufferUpdate`, `SetColorMapEntries`, `Bell`, `ServerCutText`.
- Encodings: `encodings.go` defines the `Encoding` interface and implementations like `RawEncoding`. To add one: implement `Encoding` (Type, Read, Marshal), add a constant in `encodings/encodings.go`, and ensure `ClientConn.Encodable` can return it (include in `ClientConn.encodings`).
- Messages enums: `messages/messages.go` defines wire message ids, used across the codebase.

## Conventions and patterns specific to this repo
- Files mirror RFC sections; wire structs embed padding fields to match on-the-wire layout and use big-endian via `Buffer` helpers.
- Default encodings include only Raw. If your client must handle desktop resizes, include `DesktopSizePseudoEncoding` in `ClientConn.encodings` before calling `SetEncodings`.
- Logging uses `logging.V(level) && logging.Infof(...)` patterns backed by Go's slog. Treat logs as optional; do not introduce mandatory flag parsing in library code. Configure with `logging.SetVerbosity(level)` and optionally provide a custom slog logger via `logging.SetLogger(...)`.
- A small UI settle delay is applied after client input (`KeyEvent`, `PointerEvent`, `ClientCutText`). Tests disable it via `SetSettle(0)`.
- Context tuning: the `Connect` path honors ctx value `"vnc_max_proto_version"` with values "3.3" or "3.8".

## Developer workflows
- Build/test (modules): `go test ./...` from repo root. No external services required; tests use an in-memory `MockConn`.
- Examples: See `doc.go` for a minimal client that dials a TCP VNC server, calls `Connect`, and then uses `ListenAndHandle` plus periodic `FramebufferUpdateRequest`.
- Adding a server message: implement `ServerMessage` (Type, Read). Add an instance to `ClientConfig.ServerMessages`; `ListenAndHandle` dispatches by Type().
- Adding a client message: follow the pattern in `client.go`—define a wire struct with explicit field sizes and call `c.send(...)`.

## Gotchas (repo-specific)
- Metrics accounting in `send/receive` is approximate; don’t rely on it for exact byte counts.
- If you need alpha in images, note `colorsToImage` sets alpha to 1 (nearly transparent). Adjust if you depend on visibility.
- Keep wire-level structs unexported when possible; expose higher-level helpers for users.
