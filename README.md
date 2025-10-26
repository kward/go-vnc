# VNC Library for Go
go-vnc is a VNC client library for Go.

This library implements [RFC 6143][RFC6143] -- The Remote Framebuffer Protocol
-- the protocol used by VNC.

## Project links
* CI:             [![CI][CIStatus]][CIProject]
* Documentation:  [![GoDoc][GoDocStatus]][GoDoc]

## Setup (Go modules)
1. Add to your module and run tests.

    ```
    $ go get github.com/kward/go-vnc@latest
    $ go test ./...
    ```

## Usage
Sample code usage is available in the GoDoc.

- Connect and listen to server messages: <https://pkg.go.dev/github.com/kward/go-vnc#example-Connect>
- Client input examples (keys, pointer, clipboard, updates): see [Client input examples](#client-input-examples)
- Package examples on pkg.go.dev: [keys examples](https://pkg.go.dev/github.com/kward/go-vnc/keys#pkg-examples), [buttons examples](https://pkg.go.dev/github.com/kward/go-vnc/buttons#pkg-examples)

### Client input examples

Below are small, focused examples showing how to send client event messages to a VNC server using this library. They assume you have an active connection `vc *vnc.ClientConn` from `vnc.Connect(...)`.

Key press/release (type characters or use special keys):

```go
import (
    "github.com/kward/go-vnc"
    "github.com/kward/go-vnc/keys"
)

// Press and release a printable ASCII key (using a Key constant).
_ = vc.KeyEvent(keys.A, vnc.PressKey)
_ = vc.KeyEvent(keys.A, vnc.ReleaseKey)

// Or convert from a rune for general text input.
if k, ok := keys.FromRune('!'); ok {
    _ = vc.KeyEvent(k, vnc.PressKey)
    _ = vc.KeyEvent(k, vnc.ReleaseKey)
}

// Special keys (X11 KeySyms in the 0xFFxx range) are also constants.
_ = vc.KeyEvent(keys.Return, vnc.PressKey)
_ = vc.KeyEvent(keys.Return, vnc.ReleaseKey)
```

Pointer/mouse events (move and button masks):

```go
import (
    "github.com/kward/go-vnc"
    "github.com/kward/go-vnc/buttons"
)

// Move the pointer to absolute coordinates (x, y).
_ = vc.PointerEvent(buttons.None, 100, 200)

// Left click: press then release.
_ = vc.PointerEvent(buttons.Left, 100, 200)
_ = vc.PointerEvent(buttons.None, 100, 200)

// Multiple buttons can be combined as a mask.
_ = vc.PointerEvent(buttons.Left|buttons.Right, 120, 220)
```

Send clipboard/cut text (Latin-1 only; CRs are stripped per RFC 6143 §7.5.6):

```go
import "github.com/kward/go-vnc"

// Newlines (\n) are fine; carriage returns (\r) are removed by the client.
_ = vc.ClientCutText("Line 1\r\nLine 2\n")

// Note: Only Latin-1 characters are allowed. Emojis will return an error.
if err := vc.ClientCutText("hello 😀"); err != nil {
    // handle non-Latin-1 error
}
```

Request framebuffer updates (poll or event-driven):

```go
import (
    "time"
    "github.com/kward/go-vnc"
    "github.com/kward/go-vnc/rfbflags"
)

// Periodically request incremental updates for the full desktop.
go func() {
    w, h := vc.FramebufferWidth(), vc.FramebufferHeight()
    for {
        _ = vc.FramebufferUpdateRequest(rfbflags.RFBTrue, 0, 0, w, h)
        time.Sleep(100 * time.Millisecond)
    }
}()
```

Notes:
- A small UI settle delay is applied after client input (KeyEvent, PointerEvent, ClientCutText) to avoid overwhelming remote UIs. In tests you can disable it with:

  ```go
  import "github.com/kward/go-vnc"

  vnc.SetSettle(0)
  ```

- For building text input, see helpers in `keys`:
  - `keys.FromRune(r rune) (keys.Key, bool)`
  - `keys.TextToKeys(s string) (keys.Keys, error)`
  - `keys.IntToKeys(n int) keys.Keys`

The source code is laid out such that the files match the document sections:

- [7.1] handshake.go
- [7.2] security.go
- [7.3] initialization.go
- [7.4] pixel_format.go
- [7.5] client.go
- [7.6] server.go
- [7.7] encodings.go

There are two additional files that provide everything else:

- vncclient.go -- code for instantiating a VNC client
- common.go -- common stuff not related to the RFB protocol


<!--- Links -->
[RFC6143]: http://tools.ietf.org/html/rfc6143

[CIProject]: https://github.com/kward/go-vnc/actions/workflows/go.yml
[CIStatus]: https://github.com/kward/go-vnc/actions/workflows/go.yml/badge.svg?branch=master

[GoDoc]: https://pkg.go.dev/github.com/kward/go-vnc
[GoDocStatus]: https://pkg.go.dev/badge/github.com/kward/go-vnc.svg

## Logging

This library uses a small facade over Go's slog for internal logging. You can:

- Gate verbose logs via verbosity levels:

    ```go
    import "github.com/kward/go-vnc/logging"

    // Enable result-level logs (and below). 0 disables V checks entirely.
    logging.SetVerbosity(logging.ResultLevel)
    ```

- Provide your own slog logger/handler (JSON or text) and level:

    ```go
    package main

    import (
            "log/slog"
            "os"

            "github.com/kward/go-vnc/logging"
    )

    func init() {
            logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
            logging.SetLogger(logger)
            logging.SetVerbosity(logging.ResultLevel) // optional: enable verbose gated logs
    }
    ```

- Optionally emit structured logs from your code using the facade helpers:

    ```go
    logging.Info("connecting", "addr", addr)
    logging.Debug("frame", "w", w, "h", h)
    ```

## Code generation (stringer)

This repo uses the Go stringer tool to generate String() methods for enums (e.g., Button, Key, Encoding, RFBFlag, ClientMessage, ServerMessage). To update generated files, install stringer and run go generate:

- Install stringer (Go 1.17+):

    ```bash
    go install golang.org/x/tools/cmd/stringer@latest
    ```

    Ensure your GOBIN is on PATH. By default, binaries are installed to $(go env GOPATH)/bin (or $(go env GOBIN) if set). For example, you can add this to your shell profile:

    ```bash
    export PATH="$(go env GOPATH)/bin:$PATH"
    ```

- Regenerate code from the repo root:

    ```bash
    go generate ./...
    ```

Notes:
- The go:generate directives are embedded in source files (e.g., `//go:generate stringer -type=Button`).
- Generated files have names like `*_string.go` and should be committed to the repo.

