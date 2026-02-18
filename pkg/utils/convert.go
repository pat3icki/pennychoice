package utils

import "encoding/binary"

// Convert is a namespace-like variable
var Convert = converter{}

type converter struct{}

// IntBytes converts an int to an 8-byte slice (Big Endian)
func (converter) IntBytes(v int64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(v))
	return b[:]
}
