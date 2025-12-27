package storage

import (
	"bytes"
	"os"
	"testing"
)

func TestStorage(t *testing.T) {
	tmpFile := "test_storage.dat"
	defer os.Remove(tmpFile)

	length := int64(1000)
	s, err := NewStorage(tmpFile, length)
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	defer s.Close()

	// Test Write
	data := []byte("hello world")
	offset := int64(42)

	if err := s.Write(offset, data); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Test Read
	readBuf, err := s.Read(offset, len(data))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !bytes.Equal(readBuf, data) {
		t.Errorf("Read mismatch. Got %s, Want %s", readBuf, data)
	}

	// Verify file size
	fi, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if fi.Size() != length {
		t.Errorf("File size = %d, want %d", fi.Size(), length)
	}
}
