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
