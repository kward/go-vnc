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

