package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/pat3icki/pennychoice/types"
	"golang.org/x/crypto/bcrypt"

	// Crypto Helpers
	"github.com/alexedwards/argon2id"
	scrypt "github.com/elithrar/simple-scrypt"
)

func (s *Service) ValidatePassword(ctx context.Context, usr types.User, pass string) (ValidateUser, error) {
	return s._validatePasswordOrPIN(ctx, usr, pass, false)
}

func (s *Service) ValidatePIN(ctx context.Context, usr types.User, pin string) (ValidateUser, error) {
	return s._validatePasswordOrPIN(ctx, usr, pin, true)
}

func (s *Service) ValidateAuthrazation() error

func (s *Service) ValidateNIN()

///

func (s *Service) _validatePasswordOrPIN(ctx context.Context, usr types.User, value string, isPinValue bool) (ValidateUser, error) {
	var (
		ret        ValidateUser
		invalidErr = func() (str string) {
			str = "invalid password"
			if isPinValue {
				str = "invalid pin"
			}
			return
		}
	)
	if value == "" {
		str := "password cannot be empty"
		if isPinValue {
			str = "pin cannot be empty"
		}
		return ret, errors.New(str)
	}

	userHashes, err := s._userHashes(ctx, usr)
	if err != nil {
		return ret, err
	}

	switch userHashes.HashType {
	case "bcrypt":
		// Bcrypt handles the salt/cost extraction automatically from the hash string
		err := bcrypt.CompareHashAndPassword([]byte(userHashes.HashPassword), []byte(value))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {

				return ret, errors.New(invalidErr())
			}
			return ret, err
		}
		ret = ValidateUser{
			ID:     userHashes.ID,
			Status: UserStatus(UserStatus(userHashes.Status)),
			User:   usr,
		}
		return ret, nil

	case "argon2", "argon2id":
		// Argon2id requires parsing the parameters (memory, iterations, salt) from the string.
		// The standard library (x/crypto/argon2) does not do this automatically.
		// We use the 'alexedwards/argon2id' helper to compare.
		match, err := argon2id.ComparePasswordAndHash(value, userHashes.HashPassword)
		if err != nil {
			return ret, fmt.Errorf("argon2 error: %w", err)
		}
		if !match {
			return ret, errors.New(invalidErr())
		}
		ret = ValidateUser{
			ID:     userHashes.ID,
			Status: UserStatus(userHashes.Status),
			User:   usr,
		}
		return ret, nil

	case "scrypt":
		// Similarly, scrypt needs the salt and params from the string.
		// We use 'elithrar/simple-scrypt' to compare the encoded string.
		err := scrypt.CompareHashAndPassword([]byte(userHashes.HashPassword), []byte(value))
		if err != nil {
			if errors.Is(err, scrypt.ErrMismatchedHashAndPassword) {
				return ret, errors.New(invalidErr())
			}
			return ret, fmt.Errorf("scrypt error: %w", err)
		}
		ret = ValidateUser{
			ID:     userHashes.ID,
			Status: UserStatus(userHashes.Status),
			User:   usr,
		}
		return ret, nil

	default:
		return ret, fmt.Errorf("unsupported hash type: %s", userHashes.HashType)
	}
}

func (s *Service) _userHashes(ctx context.Context, usr types.User) (UserHashes, error) {
	var (
		db_anyIdentifier sqlc.GetUserHashesRow
		ret              UserHashes
		err              error
	)

	if s.UserAnyIdentifier {
		switch usr.Uniqueness {
		case types.Unique_Email:
			db_anyIdentifier, err = s.PostgreSQL.GetUserHashes(ctx, sqlc.GetUserHashesParams{
				Phone: pgtype.Text{Valid: false},
				Email: usr.Value,
				ID:    uuid.Nil,
			})
		case types.Unique_ID:
			db_anyIdentifier, err = s.PostgreSQL.GetUserHashes(ctx, sqlc.GetUserHashesParams{
				Phone: pgtype.Text{Valid: false},
				Email: "",
				// TODO: DANGER PANIC HERE //
				ID: uuid.Must(uuid.Parse(usr.Value)),
			})
			if err != nil {
				return ret, err
			}

		case types.Unique_Phone:
			db_anyIdentifier, err = s.PostgreSQL.GetUserHashes(ctx, sqlc.GetUserHashesParams{
				Phone: pgtype.Text{String: usr.Value, Valid: true},
				Email: usr.Value,
				ID:    uuid.Nil,
			})
		default:
			return UserHashes{}, errors.New("surpported user uniquness are email, id and phone")
		}
	} else {

	}
	ret = UserHashes{
		ID:           db_anyIdentifier.ID,
		Status:       db_anyIdentifier.Status,
		HashType:     db_anyIdentifier.HashType,
		HashPassword: db_anyIdentifier.HashPassword,
		HashPin:      db_anyIdentifier.HashPin,
		HashTableSeq: uint8(db_anyIdentifier.HashTableSeq.Int16),
	}
	return ret, err
}
