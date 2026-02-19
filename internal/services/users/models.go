package users

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pat3icki/pennychoice/types"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusDeleted   UserStatus = "deleted"
	StatusSuspended UserStatus = "suspended"
)

func (u UserStatus) Is(c UserStatus) bool {
	return u == c
}

type UserProfile struct {
	FirstName   string    `json:"first_name" validate:"required,min=3,max=250"`
	MiddleName  string    `json:"middle_name,omitempty" validate:"omitempty,max=250"`
	LastName    string    `json:"last_name" validate:"required,min=3,max=250"`
	DateOfBirth time.Time `json:"date_of_birth"`
	// Runes: M=77, F=70, O=79 | m=109, f=102, o=111
	Gender          rune   `json:"gender" validate:"required,oneof=77 70 79 109 102 111"`
	Email           string `json:"email" validate:"required,email"`
	Phone           string `json:"phone" validate:"required,e164"`
	Country         string `json:"country" validate:"required,iso3166_1_alpha2"`
	IsNumberWhatApp bool   `json:"is_number_whatsapp"`
}

func (u *UserProfile) Validate(v *validator.Validate) error {
	if v == nil {
		v = validator.New()
	}
	if err := v.Struct(u); err != nil {
		return err
	}

	cutoff := time.Now().AddDate(-16, 0, 0)
	if u.DateOfBirth.After(cutoff) {
		return errors.New("date_of_birth: must be at least 16 years old")
	}
	return nil
}

type UserVParameter uint8

const (
	UserVParameter_None  UserVParameter = 0
	UserVParameter_Phone UserVParameter = 1 << iota
	UserVParameter_Email
	UserVParameter_NIN
)

func (f UserVParameter) Has(flag UserVParameter) bool {
	return f&flag != 0
}

func (f UserVParameter) Is(flag UserVParameter) bool {
	return f == flag
}

func (f UserVParameter) String() string {
	if f == 0 {
		return "None"
	}
	var parts []string
	if f.Has(UserVParameter_Phone) {
		parts = append(parts, "Phone")
	}
	if f.Has(UserVParameter_Email) {
		parts = append(parts, "Email")
	}
	if f.Has(UserVParameter_NIN) {
		parts = append(parts, "NIN")
	}
	return strings.Join(parts, "|")
}

type UserVerificationParams struct {
	Parameter UserVParameter
	Phone     bool
	Email     bool
	NIN       bool
}

type CreateUserParams struct {
	Profile      UserProfile
	Verification *UserVerificationParams
	Password     string
}

type CreateUserResult struct {
	AuthrizationID string      `json:"authorization_id"`
	Profile        UserProfile `json:"profile"`
	Status         UserStatus  `json:"status"`
}

type User struct {
	AuthrizationID string       `json:"authorization_id"`
	Profile        *UserProfile `json:"profile"`
}

type RequestKey struct {
	ID      int64     `json:"id"`
	Purpose string    `json:"purpose"`
	Period  time.Time `json:"period"`
}

type UserVerificationStatus struct {
	ID              uuid.UUID `json:"id"`
	Status          string    `json:"status"`
	IsPhoneVerified bool      `json:"is_phone_verified"`
	IsEmailVerified bool      `json:"is_email_verified"`
	IsNinVerified   bool      `json:"is_nin_verified"`
}

type UserHashes struct {
	ID           uuid.UUID `json:"id"`
	Status       string    `json:"status"`
	HashType     string    `json:"hash_type"`
	HashPassword string    `json:"hash_password"`
	HashPin      string    `json:"hash_pin"`
	HashTableSeq uint8     `json:"hash_table_seq"`
}

type ValidateUser struct {
	ID     uuid.UUID
	Status UserStatus
	User   types.User
}
