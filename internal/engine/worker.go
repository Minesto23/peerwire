package engine

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/Minesto23/peerwire/internal/peer"
	"github.com/Minesto23/peerwire/internal/piece"
	"github.com/Minesto23/peerwire/internal/tracker"
)

func (c *Client) startDownloadWorker(p tracker.Peer, workQueue chan *piece.Work, results chan *piece.Result) {
	conn, err := net.DialTimeout("tcp", p.String(), 5*time.Second)
	if err != nil {
		// fmt.Printf("Failed to connect to %s: %v\n", p, err)
		return // Exit worker
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return
	}

	// 1. Handshake
	h := peer.NewHandshake(c.InfoHash, c.PeerID)
	if err := h.Write(conn); err != nil {
		return
	}

	readH, err := peer.ReadHandshake(conn)
	if err != nil {
		return
	}

	if !bytes.Equal(readH.InfoHash[:], c.InfoHash[:]) {
		return // Wrong swarm
	}

	// 2. Initialize Peer State
	bf := make(piece.Bitfield, len(c.Spec.Info.Pieces)/20/8+1) // Roughly enough
	// Note: We don't know exact size from handshake, we wait for bitfield msg.
	// Actually bitfield message defines size.

	choked := true

	// 3. Main Loop
	for {
		// Read Message
		msg, err := peer.ReadMessage(conn)
		if err != nil {
			return // Connection closed or error
		}

		if msg == nil {
			// Keep-alive
			continue
		}

		switch msg.ID {
		case peer.MsgUnchoke:
			choked = false
		case peer.MsgChoke:
			choked = true
		case peer.MsgHave:
			index, err := peer.ParseHave(msg)
			if err == nil {
				bf.SetPiece(index)
			}
		case peer.MsgBitfield:
			// Replace our bitfield
			if len(msg.Payload) > len(bf) {
				bf = make(piece.Bitfield, len(msg.Payload))
			}
			copy(bf, msg.Payload)
		case peer.MsgPiece:
			// Handled in downloadPiece logic mostly, but if we get random piece?
			// Usually we only get pieces we requested.
		}

		// Try to work
		// Simple logic: If unchoked, try to grab work
		if !choked {
			// Pull work non-blocking or blocking?
			// Blocking is safer for the queue, but if we can't do the work, we must return it.
			select {
			case work := <-workQueue:
				if !bf.HasPiece(work.Index) {
					// Peer doesn't have it. Return to queue.
					// Potential infinite loop if all peers lack this piece.
					// Yield a bit to prevent tight loop
					workQueue <- work // Put back
					time.Sleep(time.Second)
					continue
				}

				// Download
				buf, err := downloadPiece(conn, work)
				if err != nil {
					// Failed. Put back to queue.
					workQueue <- work
					// fmt.Printf("Failed to download piece %d from %s: %v\n", work.Index, p, err)
					return // Close connection on failure to be robust (find new peer)
				}

				// Verify
				if !checkIntegrity(work, buf) {
					// Corrupt. Put back.
					workQueue <- work
					// fmt.Printf("Integrity check failed for piece %d from %s\n", work.Index, p)
					continue
				}

				// Convert to result (send copy because buf might be reused if we didn't alloc fresh)
				// We alloc fresh in downloadPiece

				// Send Have message to keep peer alive/inform
				// (Optional for leechers but good practice)

				results <- &piece.Result{Index: work.Index, Buf: buf}

			case <-time.After(time.Second):
				// Idle
			}
		} else {
			// If choked, send Interested
			// We should send interested once.
			// Simplification: Send interested always if we are choked and want data
			msg := &peer.Message{ID: peer.MsgInterested}
			msg.Write(conn)
		}
	}
}

func checkIntegrity(work *piece.Work, buf []byte) bool {
	hash := sha1.Sum(buf)
	return bytes.Equal(hash[:], work.Hash)
}

func downloadPiece(conn net.Conn, work *piece.Work) ([]byte, error) {
	// Send requests
	// Block size 16KB
	const blockSize = 16384
	// Relaxed timeout for pipelining and large pieces
	timeLeft := 60 * time.Second
	conn.SetDeadline(time.Now().Add(timeLeft))

	numBlocks := work.Length / blockSize
	if work.Length%blockSize != 0 {
		numBlocks++ // partial block at end
	}

	buf := make([]byte, work.Length)

	// Pipelining: Send all requests first, then read all responses
	// Note: For very large pieces, we might want to batch this, but strict 16KB blocks
	// means a 256KB piece is only 16 requests. This is safe to burst.

	// 1. Send ALL requests
	for i := 0; i < numBlocks; i++ {
		begin := i * blockSize
		length := blockSize
		if begin+length > work.Length {
			length = work.Length - begin
		}

		req := peer.FormatRequest(work.Index, begin, length)
		msg := &peer.Message{ID: peer.MsgRequest, Payload: req}
		if err := msg.Write(conn); err != nil {
			return nil, err
		}
	}

	// 2. Read ALL responses
	blocksReceived := 0
	for blocksReceived < numBlocks {
		resp, err := peer.ReadMessage(conn)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			continue // keep-alive
		}

		switch resp.ID {
		case peer.MsgPiece:
			// Format: <index><begin><block>
			if len(resp.Payload) < 8 {
				return nil, fmt.Errorf("piece message too short")
			}

			// Extract begin offset to know where to put it
			begin := int(binary.BigEndian.Uint32(resp.Payload[4:8]))

			if begin >= work.Length {
				return nil, fmt.Errorf("piece begin out of bounds")
			}

			data := resp.Payload[8:]
			copy(buf[begin:], data)
			blocksReceived++

		case peer.MsgChoke:
			return nil, fmt.Errorf("peer choked during download")

		case peer.MsgHave:
			// Process have message to keep bitfield up to date (optional but good)
			// For this worker function, we might just ignore or update local scope?
			// We won't update global bitfield here as we don't have access to it easily without mutex.
			// Just ignore.
			continue
		}
	}

	return buf, nil
}
