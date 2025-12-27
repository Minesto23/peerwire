package engine

import (
	"testing"

	"github.com/Minesto23/peerwire/internal/piece"
)

func TestIntegrityCheck(t *testing.T) {
	// Hash of "hello" (sha1)
	// f572d396fae9206628714fb2ce00f72e94f2258f

	hash := []byte{0xf5, 0x72, 0xd3, 0x96, 0xfa, 0xe9, 0x20, 0x66, 0x28, 0x71, 0x4f, 0xb2, 0xce, 0x00, 0xf7, 0x2e, 0x94, 0xf2, 0x25, 0x8f}

	work := &piece.Work{Hash: hash}
	buf := []byte("hello")

	if !checkIntegrity(work, buf) {
		t.Error("Integrity check failed for valid data")
	}

	bufBad := []byte("world")
	if checkIntegrity(work, bufBad) {
		t.Error("Integrity check passed for invalid data")
	}
}
