package types

type Unique uint8

const (
	Unique_ID Unique = 1 + iota
	Unique_Phone
	Unique_Email
	Unique_NIN
)

type FlagUniqueness uint8

func (f FlagUniqueness) Has(FlagUniqueness) bool

const (
	FlagUniqueness_ID = iota
	FlagUniqueness_Email
	FlagUniqueness_NIN
	FlagUniqueness_Phone
)

type User struct {
	Value      string
	Uniqueness Unique
}

func (u User) OnlyID() bool
