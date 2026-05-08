package storage

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryMarshalUnmarshal(t *testing.T) {
	original := Entry{
		Timestamp:  1715155200,
		Value:      42.5,
		MetricHash: sha256.Sum256([]byte("test_metric")),
	}
	data, err := original.MarshalBinary()

	assert.NoError(t, err)
	assert.Len(t, data, 48)

	var decoded Entry
	err = decoded.UnmarshalBinary(data)
	assert.NoError(t, err)
	assert.Equal(t, original, decoded)
}
