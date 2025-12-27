package peer

import (
	"bytes"
	"testing"
)

func TestHandshake(t *testing.T) {
	var infoHash [20]byte
	var peerID [20]byte
	copy(infoHash[:], "infohashinfohashinfo")
	copy(peerID[:], "peeridpeeridpeeridpe")

	h := NewHandshake(infoHash, peerID)

	var buf bytes.Buffer
	if err := h.Write(&buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	readH, err := ReadHandshake(&buf)
	if err != nil {
		t.Fatalf("ReadHandshake() error = %v", err)
	}

	if readH.Pstr != "BitTorrent protocol" {
		t.Errorf("Pstr = %s, want BitTorrent protocol", readH.Pstr)
	}
	if readH.InfoHash != infoHash {
		t.Errorf("InfoHash mismatch")
	}
	if readH.PeerID != peerID {
		t.Errorf("PeerID mismatch")
	}
}
