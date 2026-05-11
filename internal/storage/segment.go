package storage

import (
	"cmp"
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

// WriteSegment writes entries to a .seg file.
// Each entry is marshaled to its fixed-size EntrySize-byte binary form
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

// ScanSegments reads .seg files in dir and returns entries matching the filters.
// Files whose time range does not overlap [start, end] are skipped entirely
// without being opened — this is the block-level filtering described in ADR-003.
// Zero values for start/end or a nil hash mean "no filter".
func ScanSegments(dir string, hash *[32]byte, start, end int64) ([]Entry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", dir, err)
	}

	var result []Entry
	for _, de := range dirEntries {
		if de.IsDir() || !isSegmentFile(de.Name()) {
			continue
		}

		// Skip segments whose data range cannot overlap the query range.
		if start > 0 || end > 0 {
			segMin, segMax, err := segmentTimeRange(de.Name())
			if err != nil {
				slog.Warn("skipping segment with unparseable name", "name", de.Name(), "error", err)
				continue
			}
			if end > 0 && segMin > end {
				continue
			}
			if start > 0 && segMax < start {
				continue
			}
		}

		segPath := path.Join(dir, de.Name())
		entries, err := ReadSegment(segPath)
		if err != nil {
			slog.Warn("skipping corrupted segment", "path", segPath, "error", err)
			continue
		}

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
	}

	slices.SortFunc(result, func(a, b Entry) int {
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})

	return result, nil
}

// segmentFileName builds a segment filename encoding the data time range.
func segmentFileName(dir string, minTS, maxTS int64) string {
	return path.Join(dir, fmt.Sprintf("%d-%d.seg", minTS, maxTS))
}

// segmentTimeRange parses the min and max timestamps from a segment filename
// of the form "{minTS}-{maxTS}.seg".
func segmentTimeRange(name string) (min, max int64, err error) {
	base := strings.TrimSuffix(name, ".seg")
	parts := strings.SplitN(base, "-", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format {min}-{max}.seg, got %q", name)
	}
	min, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid min timestamp in %q: %w", name, err)
	}
	max, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid max timestamp in %q: %w", name, err)
	}
	return min, max, nil
}

// isSegmentFile returns true for files following the "{minTS}-{maxTS}.seg" naming scheme.
func isSegmentFile(name string) bool {
	if !strings.HasSuffix(name, ".seg") {
		return false
	}
	_, _, err := segmentTimeRange(name)
	return err == nil
}
