package peer

import (
	"bytes"
	"testing"
)

func TestMessage(t *testing.T) {
	// Test Have message
	// ID: 4, Payload: index 42 (4 bytes)
	// Length: 1 + 4 = 5

	payload := FormatHave(42)
	msg := &Message{ID: MsgHave, Payload: payload}

	var buf bytes.Buffer
	if err := msg.Write(&buf); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Check wire bytes manually
	// Length (4 bytes) = 5 -> 0,0,0,5
	// ID (1 byte) = 4 -> 4
	// Payload (4 bytes) = 42 -> 0,0,0,42

	expected := []byte{0, 0, 0, 5, 4, 0, 0, 0, 42}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Wire bytes mismatch. Got %v, Want %v", buf.Bytes(), expected)
	}

	readMsg, err := ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	if readMsg.ID != MsgHave {
		t.Errorf("ID = %d, want %d", readMsg.ID, MsgHave)
	}

	if !bytes.Equal(readMsg.Payload, payload) {
		t.Errorf("Payload mismatch")
	}
}

func TestKeepAlive(t *testing.T) {
	var buf bytes.Buffer
	var msg *Message = nil // Keep-alive

	msg.Write(&buf)

	expected := []byte{0, 0, 0, 0}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("KeepAlive bytes mismatch")
	}

	readMsg, err := ReadMessage(&buf)
	if err != nil {
		t.Fatalf("ReadMessage() error = %v", err)
	}

	if readMsg != nil {
		t.Errorf("Expected nil message (keep-alive), got %v", readMsg)
	}
}
