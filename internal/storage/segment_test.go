package storage

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanSegments(t *testing.T) {
	dir := t.TempDir()

	err := WriteSegment(dir+"/1000.seg", []Entry{
		{Timestamp: 100, Value: 1.0, MetricHash: sha256.Sum256([]byte("cpu"))},
		{Timestamp: 200, Value: 2.0, MetricHash: sha256.Sum256([]byte("mem"))},
	})
	require.NoError(t, err)

	got, err := ScanSegments(dir, nil, 0, 0)
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestIsSegmentFile(t *testing.T) {
	assert.True(t, isSegmentFile("1715155200.seg"))
	assert.False(t, isSegmentFile("active.wal"))
	assert.False(t, isSegmentFile("notes.txt"))
}

func TestWriteReadSegment(t *testing.T) {
	path := t.TempDir() + "/1715155200.seg"

	entries := []Entry{
		{Timestamp: 100, Value: 1.0, MetricHash: sha256.Sum256([]byte("cpu"))},
		{Timestamp: 200, Value: 2.0, MetricHash: sha256.Sum256([]byte("cpu"))},
		{Timestamp: 300, Value: 3.0, MetricHash: sha256.Sum256([]byte("mem"))},
	}

	err := WriteSegment(path, entries)
	require.NoError(t, err)

	got, err := ReadSegment(path)
	require.NoError(t, err)
	assert.Equal(t, entries, got)
}
