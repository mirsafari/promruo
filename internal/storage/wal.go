package storage

import (
	"fmt"
	"os"
	"sync"
)

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
