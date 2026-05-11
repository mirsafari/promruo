package storage

import (
	"crypto/sha256"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenWAL_CreatesFile(t *testing.T) {
	path := t.TempDir() + "/active.wal"

	wal, err := OpenWAL(path)
	require.NoError(t, err)

	assert.EqualValues(t, 0, wal.Size())

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.Size())

	err = wal.Close()
	require.NoError(t, err)
}

func TestWAL_Flusher_SortsAndPersists(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/active.wal"
	wal, err := OpenWAL(path)
	require.NoError(t, err)

	require.NoError(t, wal.Append(&Entry{Timestamp: 3, Value: 30, MetricHash: sha256.Sum256([]byte("cpu"))}))
	require.NoError(t, wal.Append(&Entry{Timestamp: 1, Value: 10, MetricHash: sha256.Sum256([]byte("cpu"))}))
	require.NoError(t, wal.Append(&Entry{Timestamp: 2, Value: 20, MetricHash: sha256.Sum256([]byte("mem"))}))

	require.NoError(t, wal.Flusher())

	// Verify segment file was created
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 2) // active.wal + one .seg file

	var segFile string
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 4 && e.Name()[len(e.Name())-4:] == ".seg" {
			segFile = dir + "/" + e.Name()
			break
		}
	}
	require.NotEmpty(t, segFile, "expected a .seg file")

	got, err := ReadSegment(segFile)
	require.NoError(t, err)
	require.Len(t, got, 3)

	// Verify sorted by timestamp
	assert.Equal(t, int64(1), got[0].Timestamp)
	assert.Equal(t, int64(2), got[1].Timestamp)
	assert.Equal(t, int64(3), got[2].Timestamp)

	wal.Close()
}

func TestWAL_AppendIncreasesSize(t *testing.T) {
	path := t.TempDir() + "/active.wal"
	wal, err := OpenWAL(path)
	require.NoError(t, err)

	err = wal.Append(&Entry{
		Timestamp:  100,
		Value:      1.5,
		MetricHash: sha256.Sum256([]byte("cpu")),
	})
	require.NoError(t, err)
	assert.EqualValues(t, 48, wal.Size())

	err = wal.Append(&Entry{
		Timestamp:  200,
		Value:      2.5,
		MetricHash: sha256.Sum256([]byte("mem")),
	})
	require.NoError(t, err)
	assert.EqualValues(t, 96, wal.Size())

	err = wal.Close()
	require.NoError(t, err)
}

func TestWAL_ReadAll(t *testing.T) {
	path := t.TempDir() + "/active.wal"
	wal, err := OpenWAL(path)
	require.NoError(t, err)

	entries := []Entry{
		{Timestamp: 1, Value: 10, MetricHash: sha256.Sum256([]byte("cpu"))},
		{Timestamp: 2, Value: 20, MetricHash: sha256.Sum256([]byte("mem"))},
	}
	for i := range entries {
		require.NoError(t, wal.Append(&entries[i]))
	}

	got, err := wal.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, entries, got)

	err = wal.Close()
	require.NoError(t, err)
}

func TestWAL_ReopenPreservesData(t *testing.T) {
	path := t.TempDir() + "/active.wal"

	wal1, err := OpenWAL(path)
	require.NoError(t, err)

	entry := &Entry{
		Timestamp:  300,
		Value:      3.14,
		MetricHash: sha256.Sum256([]byte("disk")),
	}
	err = wal1.Append(entry)
	require.NoError(t, err)

	err = wal1.Close()
	require.NoError(t, err)

	wal2, err := OpenWAL(path)
	require.NoError(t, err)

	assert.EqualValues(t, 48, wal2.Size())

	err = wal2.Close()
	require.NoError(t, err)
}
