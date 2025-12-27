package peer

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MessageID uint8

const (
	MsgChoke         MessageID = 0
	MsgUnchoke       MessageID = 1
	MsgInterested    MessageID = 2
	MsgNotInterested MessageID = 3
	MsgHave          MessageID = 4
	MsgBitfield      MessageID = 5
	MsgRequest       MessageID = 6
	MsgPiece         MessageID = 7
	MsgCancel        MessageID = 8
)

// Message represents a peer logic message (after handshake).
// Format: <4 bytes length><1 byte ID><payload>
type Message struct {
	ID      MessageID
	Payload []byte
}

// ReadMessage reads a message from the stream.
// Handles keep-alives (length 0) by returning nil, nil.
func ReadMessage(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// Keep-alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	if _, err := io.ReadFull(r, messageBuf); err != nil {
		return nil, err
	}

	return &Message{
		ID:      MessageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}, nil
}

// Write writes the message to the stream.
func (m *Message) Write(w io.Writer) error {
	if m == nil {
		// Keep-alive
		return binary.Write(w, binary.BigEndian, uint32(0))
	}

	length := uint32(1 + len(m.Payload)) // ID (1) + Payload

	// Write length
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}

	// Write ID
	if err := binary.Write(w, binary.BigEndian, m.ID); err != nil {
		return err
	}

	// Write Payload
	if len(m.Payload) > 0 {
		if _, err := w.Write(m.Payload); err != nil {
			return err
		}
	}

	return nil
}

// Helper methods for formatting payload

func FormatRequest(index, begin, length int) []byte {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return payload
}

func FormatHave(index int) []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return payload
}

func ParseHave(msg *Message) (int, error) {
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("have message expected payload 4 bytes, got %d", len(msg.Payload))
	}
	index := binary.BigEndian.Uint32(msg.Payload)
	return int(index), nil
}
