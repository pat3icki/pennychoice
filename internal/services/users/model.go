package users

import "time"

type UserStatus string

const (
	tatusActive     UserStatus = "active"
	StatusDeleted   UserStatus = "deleted"
	StatusSuspended UserStatus = "suspended"
)

func (u UserStatus) Is(c UserStatus) bool {
	return u == c
}

type UserProfile struct {
	FirstName   string    `json:"first_name"`
	MiddleName  string    `json:"middle_name,omitempty"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      rune      `json:"gender"`
}

type UserVParameter uint8

const (
	UserVParameter_None UserVParameter = iota
	UserVParameter_Phone
	UserVParameter_Email
	UserVParameter_NIN
)

type UserVerificationParams struct {
	Parameter UserVParameter
	Phone     bool
	Email     bool
	NIN       bool
}

type CreateUserParams struct {
	Profile        UserProfile
	Email          string
	Phone          string
	IsPhoneWhatApp bool
	Verification   UserVerificationParams
	HashType       string
	Password       string
	PIN            string
	NIN            string
}

type CreateUserResult struct {
	AuthrizationID  string
	Profile         UserProfile
	Status          UserStatus
	Email           string
	Phone           string
	IsPhoneWhatsApp bool
	// Verification
	HashPassword string
	HashPIN      string
}

type LoginUserParams struct {
	ValueType       uint8
	Value           string
	Password        string
	ResquestProfile bool
}

type LoginUserResult struct {
	AuthrizationID string
	Profile        *UserProfile
}

type ResquestPIN struct {
	AuthrizationID string
}
type ResquestPINResult struct {
	RpinID string
}
