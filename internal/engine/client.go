package engine

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/Minesto23/peerwire/internal/piece"
	"github.com/Minesto23/peerwire/internal/storage"
	"github.com/Minesto23/peerwire/internal/torrent"
	"github.com/Minesto23/peerwire/internal/tracker"
)

// Client is the BitTorrent client.
type Client struct {
	Spec     *torrent.TorrentSpec
	PeerID   [20]byte
	InfoHash [20]byte

	Params ClientParams
}

type ClientParams struct {
	OutputPath string
}

func NewClient(spec *torrent.TorrentSpec, params ClientParams) (*Client, error) {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return nil, err
	}
	// Prefix strict azureus style or just random? Random is fine for now,
	// but typically -PC0001- prefix.
	copy(peerID[0:8], "-PW0001-")

	return &Client{
		Spec:     spec,
		PeerID:   peerID,
		InfoHash: spec.InfoHash,
		Params:   params,
	}, nil
}

// Download starts the download process.
func (c *Client) Download(progressCb func(int, int)) error {
	// 1. Setup Storage
	store, err := storage.NewStorage(c.Params.OutputPath, c.Spec.Info.Length)
	if err != nil {
		return err
	}
	defer store.Close()

	// 2. Get Peers from Tracker
	var peers []tracker.Peer

	// Build list of all tracker URLs to try
	// Start with main announce, then flatten announce-list
	trackers := []string{c.Spec.Announce}
	for _, tier := range c.Spec.AnnounceList {
		trackers = append(trackers, tier...)
	}

	// Remove duplicates (simple O(N^2) fine for small N)
	uniqueTrackers := make([]string, 0, len(trackers))
	seen := make(map[string]bool)
	for _, tr := range trackers {
		if !seen[tr] {
			uniqueTrackers = append(uniqueTrackers, tr)
			seen[tr] = true
		}
	}

	fmt.Printf("Attempting to connect to %d trackers...\n", len(uniqueTrackers))

	for _, tr := range uniqueTrackers {
		fmt.Printf("Contacting tracker: %s\n", tr)
		foundPeers, err := tracker.RequestPeers(tr, c.InfoHash, c.PeerID, 6881, c.Spec.Info.Length)
		if err != nil {
			fmt.Printf("Tracker failed: %v\n", err)
			continue
		}
		peers = foundPeers
		fmt.Printf("Found %d peers from %s\n", len(peers), tr)
		if len(peers) > 0 {
			break // Found peers!
		}
	}

	if len(peers) == 0 {
		return fmt.Errorf("failed to find peers from any tracker")
	}

	// 3. Setup Work Queue
	workQueue := make(chan *piece.Work, len(c.Spec.Info.Pieces)/20)
	results := make(chan *piece.Result)

	// Split pieces into work items
	for index := 0; index < len(c.Spec.Info.Pieces)/20; index++ {
		hashStart := index * 20
		hashEnd := hashStart + 20
		hash := c.Spec.Info.Pieces[hashStart:hashEnd]

		length := c.Spec.Info.PieceLength
		// Last piece might be shorter
		if index == (len(c.Spec.Info.Pieces)/20)-1 {
			length = c.Spec.Info.Length % c.Spec.Info.PieceLength
			if length == 0 {
				length = c.Spec.Info.PieceLength
			}
		}

		workQueue <- &piece.Work{
			Index:  index,
			Hash:   []byte(hash), // Copy it
			Length: int(length),
		}
	}

	// 4. Start Workers
	// For this implementation, let's spawn a goroutine for each peer that was found
	for _, p := range peers {
		// Supervisor Loop: Keep reconnecting to this peer
		go func(peer tracker.Peer) {
			for {
				c.startDownloadWorker(peer, workQueue, results)
				// If returns, connection died. Wait and retry.
				// We check if download is complete to stop retrying?
				// The results channel loop checks for totalPieces.
				// But this goroutine doesn't know when to stop easily unless we close workQueue or use a context.
				// For now, simpler: retry forever. If download completes, getting work from queue will fail/block?
				// Actually, if download completes, main loop exits, program exits. So infinite retry is fine for CLI/GUI lifecycle.
				time.Sleep(10 * time.Second)
			}
		}(p)
	}

	// 5. Collect Results
	donePieces := 0
	totalPieces := len(c.Spec.Info.Pieces) / 20

	// Initial progress report
	if progressCb != nil {
		progressCb(donePieces, totalPieces)
	}

	for donePieces < totalPieces {
		res := <-results
		// Write to storage
		offset := int64(res.Index) * c.Spec.Info.PieceLength
		if err := store.Write(offset, res.Buf); err != nil {
			fmt.Printf("Error writing piece %d: %v\n", res.Index, err)
			// Should put back in queue?
			continue
		}
		donePieces++

		if progressCb != nil {
			progressCb(donePieces, totalPieces)
		}
	}

	return nil
}
