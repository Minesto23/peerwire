package torrent

import (
	"crypto/sha1"
	"errors"
	"io"

	"github.com/Minesto23/peerwire/internal/bencode"
)

// InfoDictionary represents the static metadata of the torrent.
// This is suitable for single-file torrents (as per requirements).
type InfoDictionary struct {
	PieceLength int64  `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Name        string `bencode:"name"`
	Length      int64  `bencode:"length"`
}

// TorrentSpec represents the contents of a .torrent file.
type TorrentSpec struct {
	Announce     string         `bencode:"announce"`
	AnnounceList [][]string     `bencode:"announce-list"`
	Info         InfoDictionary `bencode:"info"`
	InfoHash     [20]byte
}

// Parse reads a .torrent file and returns a TorrentSpec.
func Parse(r io.Reader) (*TorrentSpec, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// 1. Unmarshal into a raw map to handle specific structures
	//    We can't use a struct directly efficiently because our Unmarshal is basic
	//    and doesn't support struct tags yet (we skipped that complexity in Bencode).
	//    So we will manually map the map[string]interface{} to our struct.
	var raw interface{}
	if err := bencode.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, errors.New("torrent: invalid file format (not a dictionary)")
	}

	spec := &TorrentSpec{}

	// Extract Announce
	announce, ok := rawMap["announce"].(string)
	if !ok {
		return nil, errors.New("torrent: announce URL missing or invalid")
	}
	spec.Announce = announce

	// Extract Announce List (Optional)
	if tierListRaw, ok := rawMap["announce-list"].([]interface{}); ok {
		for _, tierRaw := range tierListRaw {
			if tier, ok := tierRaw.([]interface{}); ok {
				var urlTier []string
				for _, urlRaw := range tier {
					if u, ok := urlRaw.(string); ok {
						urlTier = append(urlTier, u)
					}
				}
				if len(urlTier) > 0 {
					spec.AnnounceList = append(spec.AnnounceList, urlTier)
				}
			}
		}
	}

	// Extract Info Dictionary
	infoRaw, ok := rawMap["info"]
	if !ok {
		return nil, errors.New("torrent: info dictionary missing")
	}

	infoMap, ok := infoRaw.(map[string]interface{})
	if !ok {
		return nil, errors.New("torrent: info field is not a dictionary")
	}

	// Map Info Dictionary
	spec.Info.Name, _ = infoMap["name"].(string)

	// 'length' is required for single-file mode
	if length, ok := infoMap["length"].(int64); ok {
		spec.Info.Length = length
	} else {
		return nil, errors.New("torrent: single-file 'length' missing (multi-file not supported)")
	}

	if pl, ok := infoMap["piece length"].(int64); ok {
		spec.Info.PieceLength = pl
	} else {
		return nil, errors.New("torrent: piece length missing")
	}

	if pieces, ok := infoMap["pieces"].(string); ok {
		spec.Info.Pieces = pieces
	} else {
		return nil, errors.New("torrent: pieces missing")
	}

	if len(spec.Info.Pieces)%20 != 0 {
		return nil, errors.New("torrent: pieces length not divisible by 20")
	}

	// 2. Compute InfoHash
	//    We must re-encode the *original* information found in the 'info' key.
	//    Because our map was already "dirty" (converted to Go types), re-marshaling
	//    must be done carefully. Our Bencode Encoder sorts keys, so if we pass the
	//    original map subset, it should match the canonical InfoHash.

	infoBytes, err := bencode.Marshal(infoRaw)
	if err != nil {
		return nil, err
	}

	hash := sha1.Sum(infoBytes)
	spec.InfoHash = hash

	return spec, nil
}
