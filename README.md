# HTTP from TCP (Go)

A minimal, from-scratch HTTP/1.1 server and proxy built directly on top of raw TCP sockets in Go. It focuses on practical protocol fundamentals—stream parsing, correct message framing, and streaming responses—without relying on Go’s `net/http` server stack.

## Why this exists

When you build HTTP yourself, you stop thinking in “strings and structs” and start thinking in **bytes on a connection**: partial reads, delimiters, state machines, and explicit response framing. This repo is a compact reference implementation you can read end-to-end and extend.

## Highlights

- **HTTP/1.1 request parsing** (streaming, incremental)
  - Request line: method, target, version
  - Header parsing with validation + normalization (case-insensitive keys)
  - Body parsing via `Content-Length`
- **Response writing toolkit**
  - Status line + headers + body with order enforcement
  - Default headers helper (`Content-Length`, `Connection: close`, `Content-Type`)
- **Chunked transfer encoding**
  - Streams upstream responses chunk-by-chunk (hex chunk sizes)
  - Supports **trailers** (e.g., SHA-256 + final length computed after streaming)
- **Reverse proxy endpoint**
  - `/httpbin/*` forwards to `https://httpbin.org/*` and streams the response back
- **Binary response support**
  - `/video` serves an MP4 from disk with `Content-Type: video/mp4`
- **Graceful shutdown**
  - Clean stop on SIGINT/SIGTERM

## Project layout

```text
cmd/
  httpserver/      # Main server binary
  tcplistener/     # TCP-only listener used for early debugging/inspection
  udpsender/       # UDP sender demo (helps compare TCP vs UDP behavior)
internal/
  server/          # Listener accept loop + connection handling
  request/         # Streaming request parser (state machine)
  headers/         # Header parsing + normalization utilities
  response/        # Response Writer (status/headers/body/chunked/trailers)
```

## Getting started

Requirements:
- Go 1.22+

Run the server:

```bash
go run ./cmd/httpserver
```

Try a few endpoints:

```bash
curl -v http://localhost:42069/
curl -v http://localhost:42069/yourproblem
curl -v http://localhost:42069/myproblem
curl -v http://localhost:42069/httpbin/html
curl -v http://localhost:42069/httpbin/stream/10
```

### Inspect raw chunking (recommended)

Some clients hide chunk boundaries, so use netcat to see the raw wire format:

```bash
echo -e "GET /httpbin/stream/10 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069
```

### Optional: serve the demo video

Download the video file:

```bash
mkdir -p assets
curl -o assets/vim.mp4 https://storage.googleapis.com/qvault-webapp-dynamic-assets/lesson_videos/vim-vs-neovim-prime.mp4
```

Make sure `assets/` stays untracked (recommended in `.gitignore`):

```text
assets/
```

Then open:

- http://localhost:42069/video

## Testing

```bash
go test ./...
```

## Notes

This repo was originally built as part of a guided learning track, but it’s intentionally shaped like a small “real” project: clear `cmd/` entrypoints, private `internal/` packages, focused modules, and tests around the tricky parts (parsing and framing). It’s a solid base for extending into routing, keep-alive, HTTP/2 concepts, or more robust proxying.

## License

MIT
