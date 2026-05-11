package storage

import (
	"cmp"
	"fmt"
	"os"
	"path"
	"slices"
	"sync"
	"sync/atomic"
)

const WALMaxSize = 10 << 20 // 10 MB

func sortTimestamps(data []Entry) []Entry {
	slices.SortFunc(data, func(a, b Entry) int {
		return cmp.Compare(a.Timestamp, b.Timestamp)
	})
	return data
}

type WAL struct {
	f           *os.File
	path        string
	currentSize atomic.Int64
	mu          sync.Mutex
}

func OpenWAL(path string) (*WAL, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to get file size for %s: %w", path, err)
	}

	w := &WAL{
		f:    file,
		path: path,
	}
	w.currentSize.Store(info.Size())
	return w, nil
}

func (w *WAL) Append(e *Entry) error {
	data, err := e.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal entry to binary: %w", err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	bytesWritten, err := w.f.WriteAt(data, w.currentSize.Load())
	if err != nil {
		return fmt.Errorf("failed to append entry to wal: %w", err)
	}
	w.currentSize.Add(int64(bytesWritten))

	if w.currentSize.Load() >= WALMaxSize {
		return w.flush()
	}
	return nil
}

func (w *WAL) Size() int64 {
	return w.currentSize.Load()
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

func (w *WAL) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flush()
}

// flush assumes w.mu is held by the caller.
func (w *WAL) flush() error {
	// reuse the open fd instead of opening a second one via ReadAll
	if _, err := w.f.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek wal %s: %w", w.path, err)
	}
	data, err := readEntries(w.f)
	if err != nil {
		return fmt.Errorf("failed to read from wal %s: %w", w.path, err)
	}

	sorted := sortTimestamps(data)

	if len(sorted) == 0 {
		return nil
	}

	segmentName := segmentFileName(path.Dir(w.path), sorted[0].Timestamp, sorted[len(sorted)-1].Timestamp)
	if err = WriteSegment(segmentName, sorted); err != nil {
		return fmt.Errorf("failed to write segment %s: %w", segmentName, err)
	}

	if err := w.f.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate wal %s: %w", w.path, err)
	}
	if _, err := w.f.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek wal %s: %w", w.path, err)
	}
	w.currentSize.Store(0)

	return nil
}


