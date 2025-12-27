package peer

import (
	"fmt"
	"io"
)

// Handshake is a special message used to establish a connection.
type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// NewHandshake creates a new handshake with the standard protocol string.
func NewHandshake(infoHash [20]byte, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// Write writes the handshake to w.
func (h *Handshake) Write(w io.Writer) error {
	buf := make([]byte, 68)
	buf[0] = byte(len(h.Pstr))
	copy(buf[1:20], []byte(h.Pstr))
	// Bytes 20-28 are reserved (zeros), inherently zero in make slice
	copy(buf[28:48], h.InfoHash[:])
	copy(buf[48:68], h.PeerID[:])

	_, err := w.Write(buf)
	return err
}

// Read reads a handshake from r.
func ReadHandshake(r io.Reader) (*Handshake, error) {
	buf := make([]byte, 68) // Handshake is exactly 68 bytes

	// We read full 68 bytes. In a real scenario we might read length first to check protocol,
	// but standard handshake is fixed length for "BitTorrent protocol" (19 chars + 1 len = 20) + 8 reserved + 20 hash + 20 id = 68.
	// Actually pstrlen is 1 byte.
	// 1 + 19 + 8 + 20 + 20 = 68.

	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	pstrlen := int(buf[0])
	if pstrlen == 0 {
		return nil, fmt.Errorf("invalid handshake: pstrlen is 0")
	}

	// We only strictly support "BitTorrent protocol" for now
	pstr := string(buf[1 : 1+pstrlen])
	if pstr != "BitTorrent protocol" {
		return nil, fmt.Errorf("unknown protocol: %s", pstr)
	}

	var infoHash [20]byte
	var peerID [20]byte

	// Reserved bytes at buf[1+pstrlen : 1+pstrlen+8] -> ignore

	copy(infoHash[:], buf[1+pstrlen+8:1+pstrlen+8+20])
	copy(peerID[:], buf[1+pstrlen+8+20:1+pstrlen+8+20+20])

	return &Handshake{
		Pstr:     pstr,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
}
