package storage

import (
	"os"
)

// Storage handles reading and writing to the target file.
type Storage struct {
	file *os.File
}

// NewStorage opens or creates the file at path with the given length.
func NewStorage(path string, length int64) (*Storage, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	// Truncate to ensure correct size
	if err := file.Truncate(length); err != nil {
		file.Close()
		return nil, err
	}

	return &Storage{file: file}, nil
}

// Write writes a block of data at the specified offset.
func (s *Storage) Write(offset int64, data []byte) error {
	_, err := s.file.WriteAt(data, offset)
	return err
}

// Read reads a block of data from the specified offset.
func (s *Storage) Read(offset int64, length int) ([]byte, error) {
	buf := make([]byte, length)
	_, err := s.file.ReadAt(buf, offset)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Close closes the file.
func (s *Storage) Close() error {
	return s.file.Close()
}
