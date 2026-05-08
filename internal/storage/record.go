package storage

import (
	"encoding/binary"
	"errors"
	"math"
)

type Entry struct {
	Timestamp  int64
	Value      float64
	MetricHash [32]byte // fixed-size array for SHA-256 output
}

func (e *Entry) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 48)
	binary.LittleEndian.PutUint64(buf[0:8], uint64(e.Timestamp))
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(e.Value))
	copy(buf[16:48], e.MetricHash[:]) // copy requires a slice not an array, so [:] conversion is needed

	return buf, nil
}

func (e *Entry) UnmarshalBinary(d []byte) error {
	if len(d) != 48 {
		return errors.New("byte array too long or too short")
	}

	t := binary.LittleEndian.Uint64(d[0:8])
	v := math.Float64frombits(binary.LittleEndian.Uint64(d[8:16]))
	var mh [32]byte
	copy(mh[:], d[16:48])

	e.Timestamp = int64(t)
	e.Value = v
	e.MetricHash = mh
	return nil
}
