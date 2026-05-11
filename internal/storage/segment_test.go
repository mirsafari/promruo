package storage

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanSegments(t *testing.T) {
	dir := t.TempDir()

	err := WriteSegment(dir+"/100-200.seg", []Entry{
		{Timestamp: 100, Value: 1.0, MetricHash: sha256.Sum256([]byte("cpu"))},
		{Timestamp: 200, Value: 2.0, MetricHash: sha256.Sum256([]byte("mem"))},
	})
	require.NoError(t, err)

	got, err := ScanSegments(dir, nil, 0, 0)
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestScanSegments_SkipsNonOverlapping(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, WriteSegment(dir+"/100-200.seg", []Entry{
		{Timestamp: 100, Value: 1.0, MetricHash: sha256.Sum256([]byte("cpu"))},
		{Timestamp: 200, Value: 2.0, MetricHash: sha256.Sum256([]byte("cpu"))},
	}))
	require.NoError(t, WriteSegment(dir+"/500-600.seg", []Entry{
		{Timestamp: 500, Value: 5.0, MetricHash: sha256.Sum256([]byte("cpu"))},
		{Timestamp: 600, Value: 6.0, MetricHash: sha256.Sum256([]byte("cpu"))},
	}))

	// query range [150, 300] should only match the first segment
	got, err := ScanSegments(dir, nil, 150, 300)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(200), got[0].Timestamp)
}

func TestIsSegmentFile(t *testing.T) {
	assert.True(t, isSegmentFile("1715155200-1715758800.seg"))
	assert.False(t, isSegmentFile("1715155200.seg")) // old single-timestamp format
	assert.False(t, isSegmentFile("active.wal"))
	assert.False(t, isSegmentFile("notes.txt"))
}

func TestWriteReadSegment(t *testing.T) {
	path := t.TempDir() + "/100-300.seg"

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

func TestSegmentTimeRange(t *testing.T) {
	min, max, err := segmentTimeRange("1715155200-1715758800.seg")
	require.NoError(t, err)
	assert.Equal(t, int64(1715155200), min)
	assert.Equal(t, int64(1715758800), max)

	_, _, err = segmentTimeRange("notvalid.seg")
	assert.Error(t, err)
}
