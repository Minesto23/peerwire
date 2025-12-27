package piece

// Work represents a piece to be downloaded.
type Work struct {
	Index  int
	Hash   []byte
	Length int
}

// Result represents a downloaded piece.
type Result struct {
	Index int
	Buf   []byte
}
