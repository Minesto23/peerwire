package tracker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"
)

// UDP Protocol Constants
const (
	protocolId     = 0x41727101980
	actionConnect  = 0
	actionAnnounce = 1
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func requestPeersUDP(announceURL string, infoHash [20]byte, peerID [20]byte, port int, length int64) ([]Peer, error) {
	var lastErr error
	// Try up to 3 times
	for i := 0; i < 3; i++ {
		peers, err := doRequestPeersUDP(announceURL, infoHash, peerID, port, length)
		if err == nil {
			return peers, nil
		}
		lastErr = err
		// Exponential backoff or just short wait?
		// BEP 15 suggests 15 * 2^n. But user might be impatient.
		// Let's try 5s, 10s, 15s logic inside the attempt, or just close and retry.
		// We'll just retry immediately with a fresh socket, assuming packet loss.
		fmt.Printf("UDP Tracker Attempt %d/3 failed: %v. Retrying...\n", i+1, err)
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("udp tracker failed after 3 attempts: %v", lastErr)
}

func doRequestPeersUDP(announceURL string, infoHash [20]byte, peerID [20]byte, port int, length int64) ([]Peer, error) {
	parsed, err := url.Parse(announceURL)
	if err != nil {
		return nil, err
	}

	serverAddr, err := net.ResolveUDPAddr("udp", parsed.Host)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set a reasonable deadline for each attempt
	// 5 seconds for connect + announce is tight but responsive.
	// BEP 15 suggests 15s. Let's use 10s per attempt.
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// 1. Connection Request
	transactionID := rand.Uint32()

	connectReq := new(bytes.Buffer)
	binary.Write(connectReq, binary.BigEndian, uint64(protocolId))
	binary.Write(connectReq, binary.BigEndian, uint32(actionConnect))
	binary.Write(connectReq, binary.BigEndian, uint32(transactionID))

	if _, err := conn.Write(connectReq.Bytes()); err != nil {
		return nil, err
	}

	// Connection Response (16 bytes)
	connRespBuf := make([]byte, 16)
	n, err := conn.Read(connRespBuf)
	if err != nil {
		return nil, err
	}
	if n < 16 {
		return nil, errors.New("udp tracker: connection response too short")
	}

	action := binary.BigEndian.Uint32(connRespBuf[0:4])
	tid := binary.BigEndian.Uint32(connRespBuf[4:8])
	connID := binary.BigEndian.Uint64(connRespBuf[8:16])

	if action != actionConnect {
		return nil, fmt.Errorf("udp tracker: connect action mismatch, got %d", action)
	}
	if tid != transactionID {
		return nil, fmt.Errorf("udp tracker: connect transaction id mismatch")
	}

	// 2. Announce Request
	// Offset  Size    Name
	// 0       8       connection_id
	// 8       4       action = 1
	// 12      4       transaction_id
	// 16      20      info_hash
	// 36      20      peer_id
	// 56      8       downloaded
	// 64      8       left
	// 72      8       uploaded
	// 80      4       event
	// 84      4       IP address
	// 88      4       key
	// 92      4       num_want
	// 96      2       port

	transactionID = rand.Uint32()

	announceReq := new(bytes.Buffer)
	binary.Write(announceReq, binary.BigEndian, uint64(connID))
	binary.Write(announceReq, binary.BigEndian, uint32(actionAnnounce))
	binary.Write(announceReq, binary.BigEndian, uint32(transactionID))
	announceReq.Write(infoHash[:])
	announceReq.Write(peerID[:])
	binary.Write(announceReq, binary.BigEndian, uint64(0))             // downloaded
	binary.Write(announceReq, binary.BigEndian, uint64(length))        // left
	binary.Write(announceReq, binary.BigEndian, uint64(0))             // uploaded
	binary.Write(announceReq, binary.BigEndian, uint32(0))             // event: none
	binary.Write(announceReq, binary.BigEndian, uint32(0))             // ip: default
	binary.Write(announceReq, binary.BigEndian, uint32(rand.Uint32())) // key
	binary.Write(announceReq, binary.BigEndian, int32(-1))             // num_want: default
	binary.Write(announceReq, binary.BigEndian, uint16(port))

	if _, err := conn.Write(announceReq.Bytes()); err != nil {
		return nil, err
	}

	// Announce Response
	// action (4), trans_id (4), interval (4), leechers (4), seeders (4), peers...
	respBuf := make([]byte, 4096)
	n, err = conn.Read(respBuf)
	if err != nil {
		return nil, err
	}

	if n < 20 {
		return nil, errors.New("udp tracker: announce response too short")
	}

	action = binary.BigEndian.Uint32(respBuf[0:4])
	if action != actionAnnounce {
		return nil, fmt.Errorf("udp tracker: announce action mismatch, got %d", action)
	}

	// Parses peers from remaining bytes
	// Each peer is 6 bytes (IP 4, Port 2)
	peersBin := respBuf[20:n]

	var peers []Peer
	for i := 0; i < len(peersBin)/6; i++ {
		offset := i * 6
		ip := net.IP(peersBin[offset : offset+4])
		port := binary.BigEndian.Uint16(peersBin[offset+4 : offset+6])
		peers = append(peers, Peer{IP: ip, Port: port})
	}

	return peers, nil
}
