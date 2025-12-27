package piece

import "testing"

func TestBitfield(t *testing.T) {
	// 2 bytes = 16 bits
	bf := make(Bitfield, 2)

	// Set 0th piece (first bit)
	bf.SetPiece(0)
	// Byte 0 should be 10000000 (0x80)
	if bf[0] != 0x80 {
		t.Errorf("Bitfield[0] = %x, want 0x80", bf[0])
	}
	if !bf.HasPiece(0) {
		t.Errorf("HasPiece(0) = false, want true")
	}

	// Set 7th piece (last bit of first byte)
	bf.SetPiece(7)
	// Byte 0 should be 10000001 (0x81)
	if bf[0] != 0x81 {
		t.Errorf("Bitfield[0] = %x, want 0x81", bf[0])
	}
	if !bf.HasPiece(7) {
		t.Errorf("HasPiece(7) = false, want true")
	}

	// Set 8th piece (first bit of second byte)
	bf.SetPiece(8)
	// Byte 1 should be 10000000 (0x80)
	if bf[1] != 0x80 {
		t.Errorf("Bitfield[1] = %x, want 0x80", bf[1])
	}
}
