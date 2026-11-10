package utils

import "unsafe"

type _unsafe struct{}

// Create a global instance
var Unsafe = _unsafe{}

func (_unsafe) BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func (_unsafe) StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
