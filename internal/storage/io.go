package storage

import (
	"fmt"
	"io"
)

// readEntries reads all entries from any io.Reader (file, bytes.Buffer, etc.).
// Each entry is a fixed-size EntrySize-byte record, so we read in exact EntrySize-byte chunks.
// We do not use bufio.Scanner here because binary records can contain arbitrary
// byte values including newlines, which would cause a line scanner to split incorrectly.
func readEntries(r io.Reader) ([]Entry, error) {
	var entries []Entry
	buf := make([]byte, EntrySize)
	for {
		_, err := io.ReadFull(r, buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read entry: %w", err)
		}
		var entry Entry
		if err := entry.UnmarshalBinary(buf); err != nil {
			return nil, fmt.Errorf("corrupted entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
