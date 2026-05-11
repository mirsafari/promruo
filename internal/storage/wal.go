package storage

import (
	"fmt"
	"os"
	"path"
	"slices"
	"strconv"
	"sync"
	"time"
)

func sortTimestamps(data []Entry) []Entry {
	slices.SortFunc(data, func(a, b Entry) int {
		if a.Timestamp < b.Timestamp {
			return -1
		}
		if a.Timestamp > b.Timestamp {
			return 1
		}
		return 0
	})
	return data
}

type WAL struct {
	f           *os.File
	path        string
	currentSize int64
	mu          sync.Mutex
}

func OpenWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", path, err)
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to get file size for %s: %v", path, err)
	}

	return &WAL{
		f:           file,
		path:        path,
		currentSize: info.Size(),
	}, nil
}

func (w *WAL) Append(e *Entry) error {
	data, err := e.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal entry to binary: %v", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	bytesWritten, err := w.f.WriteAt(data, w.currentSize)
	if err != nil {
		return fmt.Errorf("failed to append entry to wal: %v", err)
	}
	w.currentSize += int64(bytesWritten)

	return nil
}

func (w *WAL) Size() int64 {
	return w.currentSize
}

func (w *WAL) Close() error {
	err := w.f.Sync()
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", w.path, err)
	}
	err = w.f.Close()
	if err != nil {
		return fmt.Errorf("failed to close file %s: %v", w.path, err)
	}
	return nil
}

func (w *WAL) ReadAll() ([]Entry, error) {
	file, err := os.Open(w.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open wal file %s: %w", w.path, err)
	}
	defer func() { _ = file.Close() }()

	entries, err := readEntries(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read wal %s: %w", w.path, err)
	}

	return entries, nil
}

func (w *WAL) Flusher() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := w.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read from wal %s: %w", w.path, err)
	}

	sorted := sortTimestamps(data)

	segmentDir := path.Dir(w.path)
	segmentName := path.Join(segmentDir, strconv.FormatInt(time.Now().Unix(), 10)+".seg")
	err = WriteSegment(segmentName, sorted)
	if err != nil {
		return fmt.Errorf("failed to write segment %s: %w", segmentName, err)
	}

	if err := w.f.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate wal %s: %w", w.path, err)
	}
	if _, err := w.f.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek wal %s: %w", w.path, err)
	}
	w.currentSize = 0

	return nil
}


