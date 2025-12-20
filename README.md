# peerwire

A BitTorrent client written **from scratch** in Go, using **only the standard library**.

This project is focused on learning and implementing the BitTorrent protocol
at a low level: binary parsing, networking, concurrency, and file integrity.

📘 [Protocol Overview](docs/protocol.md)


## Goals
- Zero external dependencies
- Full control over protocol implementation
- Clean, idiomatic Go architecture

## Roadmap
- [ ] Bencode decoder / encoder
- [ ] Torrent file parsing
- [ ] Info hash generation
- [ ] HTTP tracker support
- [ ] Peer handshake (Peer Wire Protocol)
- [ ] Piece scheduling & download
- [ ] File assembly and verification

## Usage
```bash
go run ./cmd/peerwire <file.torrent>
