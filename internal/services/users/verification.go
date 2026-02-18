package users

import (
	"context"
	"errors"
	"unsafe"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/pat3icki/pennychoice/types"
)

func (s *Service) Verifiy(ctx context.Context, usr types.User, ver types.FlagUniqueness) error {
	var (
		usr_verifiy sqlc.UpdateUserVerificationParams
		_is_updated bool
		id          uuid.UUID
		err         error
	)
	ret, err := s.GetVerificationStatus(ctx, usr)
	if err != nil {
		return err
	}

	if ret.Status != "active" {
		// TODO: err
		return errors.ErrUnsupported
	}
	// Email
	if ver.Has(types.FlagUniqueness_Email) {
		usr_verifiy.IsEmailVerified = true

		_is_updated = true
	}
	// Phone
	if ver.Has(types.FlagUniqueness_Phone) {
		usr_verifiy.IsPhoneVerified = true
		_is_updated = true

	}
	// NIN
	if ver.Has(types.FlagUniqueness_NIN) {
		usr_verifiy.IsNinVerified = true
		_is_updated = true

	}
	if _is_updated {
		id, err = s.PostgreSQL.UpdateUserVerification(ctx, usr_verifiy)
		if err != nil {
			return err
		}
	}
	if id != usr_verifiy.ID {
		// TODO: correct error
		return err
	}
	return nil
}

func (s *Service) GetVerificationStatus(ctx context.Context, usr types.User) (UserVerificationStatus, error) {
	var (
		usr_verify UserVerificationStatus
		id         uuid.UUID
		err        error
	)

	switch usr.Uniqueness {
	// User Email
	case types.Unique_Email:
		db_ret, err := s.PostgreSQL.GetUserVerificationByEmail(ctx, usr.Value)
		if err != nil {
			return usr_verify, err
		}
		// dont worry its safe :)
		usr_verify = *(*UserVerificationStatus)(unsafe.Pointer(&db_ret))
	// User Email
	case types.Unique_Phone:
		db_ret, err := s.PostgreSQL.GetUserVerificationByPhone(ctx, pgtype.Text{String: usr.Value, Valid: true})
		if err != nil {
			return usr_verify, err
		}
		// dont worry its safe :)
		usr_verify = *(*UserVerificationStatus)(unsafe.Pointer(&db_ret))

		// User UUID
	case types.Unique_ID:
		id, err = uuid.Parse(usr.Value)
		if err != nil {
			return usr_verify, err
		}
		db_ret, err := s.PostgreSQL.GetUserVerificationByID(ctx, id)
		if err != nil {
			return usr_verify, err
		}
		usr_verify = *(*UserVerificationStatus)(unsafe.Pointer(&db_ret))

	default:
		return usr_verify, errors.New("unsupported uniqueness type")

	}
	return usr_verify, err
}
