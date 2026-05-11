package storage

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"
)

// WriteSegment writes entries to a .seg file.
// Each entry is marshaled to its fixed-size 48-byte binary form
// and written sequentially. The caller is responsible for ensuring
// entries are sorted by timestamp.
func WriteSegment(path string, entries []Entry) error {
	// O_TRUNC ensures that if a segment file already exists it is wiped clean before writing.
	// Without this flag, leftover bytes from a larger previous write could remain
	// at the end of the file, corrupting subsequent reads.
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("failed to create segment file %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	for i := range entries {
		data, err := entries[i].MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal entry %d in segment %s: %w", i, path, err)
		}
		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("failed to write entry %d to segment %s: %w", i, path, err)
		}
	}

	return nil
}

// ReadSegment reads all entries from a .seg file.
func ReadSegment(path string) ([]Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open segment file %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	entries, err := readEntries(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read segment %s: %w", path, err)
	}

	return entries, nil
}

// ScanSegments reads all .seg files in a directory and returns entries
func ScanSegments(dir string, hash *[32]byte, start, end int64) ([]Entry, error) {
	entries, err := readAllEntries(dir)
	if err != nil {
		return nil, err
	}

	var result []Entry
	for _, e := range entries {
		if hash != nil && e.MetricHash != *hash {
			continue
		}
		if start > 0 && e.Timestamp < start {
			continue
		}
		if end > 0 && e.Timestamp > end {
			continue
		}
		result = append(result, e)
	}

	slices.SortFunc(result, func(a, b Entry) int {
		if a.Timestamp < b.Timestamp {
			return -1
		}
		if a.Timestamp > b.Timestamp {
			return 1
		}
		return 0
	})

	return result, nil
}

func readAllEntries(dir string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", dir, err)
	}

	var all []Entry
	for _, de := range dirEntries {
		if de.IsDir() || !isSegmentFile(de.Name()) {
			continue
		}
		segPath := path.Join(dir, de.Name())
		entries, err := ReadSegment(segPath)
		if err != nil {
			slog.Warn("skipping corrupted segment", "path", segPath, "error", err)
			continue
		}
		all = append(all, entries...)
	}

	return all, nil
}

func isSegmentFile(name string) bool {
	return strings.HasSuffix(name, ".seg")
}
