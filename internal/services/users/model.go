package users

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	FirstName       string    `json:"first_name"`
	MiddleName      string    `json:"middle_name,omitempty"`
	LastName        string    `json:"last_name"`
	DateOfBirth     time.Time `json:"date_of_birth"`
	Gender          rune      `json:"gender"`
	Email           string
	Phone           string
	IsNumberWhatApp bool
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
	Profile      UserProfile
	Verification *UserVerificationParams
	Password     string
}

type CreateUserResult struct {
	AuthrizationID string
	Profile        UserProfile
	Status         UserStatus
	// Verification
}

type LoginParams struct {
	User            types.User
	Password        string
	ResquestProfile bool
}

type User struct {
	AuthrizationID string
	Profile        *UserProfile
}

type ResquestPIN struct {
	AuthrizationID string
}
type ResquestPINResult struct {
	RpinID string
}

type RequestKey struct {
	ID      int64
	Purpose string
	Period  time.Time
}

type UserVerificationStatus struct {
	ID              uuid.UUID `db:"id" json:"id"`
	Status          string    `db:"status" json:"status"`
	IsPhoneVerified bool      `db:"is_phone_verified" json:"is_phone_verified"`
	IsEmailVerified bool      `db:"is_email_verified" json:"is_email_verified"`
	IsNinVerified   bool      `db:"is_nin_verified" json:"is_nin_verified"`
}

type GetUserHashes struct {
	ID           uuid.UUID   `db:"id" json:"id"`
	Status       string      `db:"status" json:"status"`
	HashType     string      `db:"hash_type" json:"hash_type"`
	HashPassword string      `db:"hash_password" json:"hash_password"`
	HashPin      string      `db:"hash_pin" json:"hash_pin"`
	HashTableSeq pgtype.Int2 `db:"hash_table_seq" json:"hash_table_seq"`
}

type ValidateUser struct {
	ID     uuid.UUID
	Status string
	User   types.User
}
