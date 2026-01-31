package utils

type ParamValue struct {
	Type  uint8
	Value string
}

type Unique uint8

const (
	Unique_ID Unique = 1 + iota
	Unique_Phone
	Unique_Email
	Unique_NIN
)

type User struct {
	Value      string
	Uniqueness Unique
}
