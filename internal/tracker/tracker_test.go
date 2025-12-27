package tracker

import (
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/Minesto23/peerwire/internal/bencode"
)

func TestRequestPeers(t *testing.T) {
    // Mock Tracker Server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify query params
        q := r.URL.Query()
        if q.Get("info_hash") == "" {
            http.Error(w, "missing info_hash", 400)
            return
        }
        
        // Return dummy response
        // Interval: 900
        // Peers: 127.0.0.1:8080 (7f000001 1f90)
        
        // Manually build bencoded data since we want strict binary control
        // d8:intervali900e5:peers6:<binary>e
        
        peersBin := []byte{127, 0, 0, 1, 0x1f, 0x90}
        
        resp := map[string]interface{}{
            "interval": 900,
            "peers":    string(peersBin), // bencode expects string for binary data in many cases
        }
        
        data, _ := bencode.Marshal(resp)
        w.Write(data)
    }))
    defer server.Close()
    
    var infoHash [20]byte
    var peerID [20]byte
    copy(infoHash[:], "12345678901234567890")
    copy(peerID[:], "peerID12345678901234")
    
    peers, err := RequestPeers(server.URL, infoHash, peerID, 6881, 1000)
    if err != nil {
        t.Fatalf("RequestPeers failed: %v", err)
    }
    
    if len(peers) != 1 {
        t.Fatalf("Expected 1 peer, got %d", len(peers))
    }
    
    expectedIP := "127.0.0.1"
    if peers[0].IP.String() != expectedIP {
        t.Errorf("IP = %s, want %s", peers[0].IP.String(), expectedIP)
    }
    
    if peers[0].Port != 8080 {
        t.Errorf("Port = %d, want 8080", peers[0].Port)
    }
}
