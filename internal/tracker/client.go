package tracker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Minesto23/peerwire/internal/bencode"
)

// Peer represents a peer retrieved from the tracker.
type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

// RequestPeers connects to a tracker and returns a list of peers.
func RequestPeers(announceURL string, infoHash [20]byte, peerID [20]byte, port int, length int64) ([]Peer, error) {
	base, err := url.Parse(announceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid announce URL: %v", err)
	}

	if base.Scheme == "udp" {
		return requestPeersUDP(announceURL, infoHash, peerID, port, length)
	}

	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.FormatInt(length, 10)},
	}

	base.RawQuery = params.Encode()

	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(base.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("tracker returned error status: %d", resp.StatusCode)
	}

	// Parse Bencoded response
	// Format: d8:intervali900e5:peers6:xxxxxx...e

	// We need to read body first
	var result map[string]interface{}
	// Note: bencode.Unmarshal signature in my previous step was taking []byte.
	// We didn't change it to io.Reader in signature, but internal decode took reader.
	// Oh, in decoder.go: Unmarshal(data []byte, v interface{})
	// So we must read body.

	// RE-READING my own code: decoder.go Unmarshal takes DATA []BYTE.

	bodyBytes, err := readAll(resp.Body) // Helper or standard io.ReadAll in Go 1.16+
	if err != nil {
		return nil, err
	}

	if err := bencode.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}

	if fail, ok := result["failure reason"].(string); ok {
		return nil, errors.New("tracker failure: " + fail)
	}

	peersRaw, ok := result["peers"].(string)
	if !ok {
		// It might be list of dicts (non-compact), but we requested compact=1
		// For now fail if not string
		return nil, errors.New("tracker peers list not a string (compact mode expected)")
	}

	return parsePeers(peersRaw)
}

func parsePeers(peersBin string) ([]Peer, error) {
	const peerSize = 6 // 4 bytes IP, 2 bytes Port
	if len(peersBin)%peerSize != 0 {
		return nil, errors.New("received malformed peers list")
	}

	numPeers := len(peersBin) / peerSize
	peers := make([]Peer, numPeers)

	for i := 0; i < numPeers; i++ {
		offset := i * peerSize

		ip := net.IP([]byte(peersBin[offset : offset+4]))
		port := binary.BigEndian.Uint16([]byte(peersBin[offset+4 : offset+6]))

		peers[i] = Peer{IP: ip, Port: port}
	}

	return peers, nil
}

// io.ReadAll alias to avoid import confusion if using old go but here accessing standard lib
func readAll(r io.Reader) ([]byte, error) {
	// using io.ReadAll
	// can just import "io"
	return io.ReadAll(r)
}
