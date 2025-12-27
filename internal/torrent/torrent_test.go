package torrent

import (
    "bytes"
    "crypto/sha1"
    "testing"

    "github.com/Minesto23/peerwire/internal/bencode"
)

func TestParse(t *testing.T) {
	// Construct a manual Bencode buffer to simulate a .torrent file
	infoDict := map[string]interface{}{
		"name":         "testfile",
		"length":       int64(12345),
		"piece length": int64(256),
		"pieces":       "12345678901234567890", // 20 bytes dummy hash
	}
	
	rootDict := map[string]interface{}{
		"announce": "http://tracker.example.com",
		"info":     infoDict,
	}
	
	data, err := bencode.Marshal(rootDict)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	
	spec, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	
	if spec.Announce != "http://tracker.example.com" {
		t.Errorf("Announce = %s, want http://tracker.example.com", spec.Announce)
	}
	
	if spec.Info.Length != 12345 {
		t.Errorf("Length = %d, want 12345", spec.Info.Length)
	}
	
    // Verify InfoHash matches SHA1(bencode(infoDict))
    infoBytes, _ := bencode.Marshal(infoDict)
    rawHash := sha1.Sum(infoBytes)
    
    if spec.InfoHash != rawHash {
        t.Errorf("InfoHash mismatch")
    }
    if len(spec.InfoHash) != 20 {
        t.Errorf("InfoHash length = %d, want 20", len(spec.InfoHash))
    }
}
