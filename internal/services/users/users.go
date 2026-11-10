package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pat3icki/pennychoice/gmodels"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"

	"github.com/pat3icki/encrypt"
	"github.com/pat3icki/pennychoice/pkg/jschema"
	"github.com/pat3icki/pennychoice/pkg/messenger"
	"github.com/pat3icki/pennychoice/pkg/utils"
)

// Manager provides user‑related services such as profile creation and login.
// It holds a reference to the database client and a messenger client for
// sending notifications.
type Service struct {
	mu          sync.Mutex
	Postgres    *db.Client
	Messenger   messenger.Client
	SchemanName string
	// Audit     audit.Auditor
}

// CreateProfile validates the incoming payload, hashes the password (and optional PIN),
// inserts a new user record into the database and writes an audit entry.
// It runs all write operations inside a transaction.
func (m *Service) CreateProfile(ctx context.Context, params *CreateUserProfileParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		_user         CreateProfile
		hash_password []byte
		id            uuid.UUID
		err           error
	)

	// Validate payload against JSON schema
	err = jschema.Validate("create_user.schema", params.Payload, nil)
	if err != nil {
		return fmt.Errorf("THERE WAS A FUCKING ERROR HERE: %W", err)
	}

	// Unmarshal payload into struct
	err = json.Unmarshal(params.Payload, &_user)
	if err != nil {
		return err
	}

	// Generate a new UUID for the user
	id, err = m.Postgres.GenerateUUID(ctx, db.TableUsers)
	if err != nil {
		return err
	}

	pass := _user.Password
	hash_password, err = encrypt.CreateHash(encrypt.Bcrypt{Cost: 12}, utils.Unsafe.StringToBytes(pass))
	if err != nil {
		return err
	}
	var hash_pin pgtype.Text

	if _user.PIN != "" {
		pin := _user.PIN
		p, err := encrypt.CreateHash(encrypt.Bcrypt{Cost: 12}, utils.Unsafe.StringToBytes(pin))
		if err != nil {
			return err
		}
		hash_pin = pgtype.Text{String: string(p), Valid: true}
	}
	return m.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {

		creat_params := sqlc.CreateUserProfileFULLParams{
			ID:              id,
			FirstName:       _user.FirstName,
			MiddleName:      utils.PgConvert.ToPgText(_user.MiddleName),
			LastName:        _user.LastName,
			DateOfBirth:     pgtype.Date{},
			Gender:          "M", // implement Constants
			Status:          "active",
			Email:           _user.Email,
			Phone:           utils.PgConvert.ToPgText(_user.Phone),
			Country:         "Nigeria",
			IsWhatsappPhone: false,
			HashPassword:    hash_password,
			HashPin:         hash_pin,
		}
		ret, err := q.CreateUserProfileFULL(ctx, creat_params)
		if err != nil {
			return err
		}
		if (ret.ID != creat_params.ID) || !bytes.Equal(ret.HashPassword, creat_params.HashPassword) {
			return errors.New("failed to create user profile")
		}

		return nil

	})

}

// CreateContestantProfile validates the incoming payload, checks that the
// profile name is not already taken for the given user, then inserts a new
// contestant-profile row and writes an audit entry — all inside a transaction.
func (m *Service) CreateContestantProfile(ctx context.Context, user uuid.UUID, params *CreateContestantProfileParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var _profile CreateContestantProfile

	// Validate payload against JSON schema
	err := jschema.Validate("create_contestant.schema", params.Payload, nil)
	if err != nil {
		return err
	}

	// Unmarshal payload into struct
	err = json.Unmarshal(params.Payload, &_profile)
	if err != nil {
		return err
	}

	// Parse the user UUID supplied in the payload
	_, err = m.Postgres.Exists(user, db.TableUsers)
	if err != nil {
		return fmt.Errorf("campaign `id` cannot be empty")
	}

	// Ensure a contestant profile with the same name does not already exist
	_, err = m.Postgres.Query.CheckUserContestantProfileExists(ctx, sqlc.CheckUserContestantProfileExistsParams{
		UserID:      user,
		ProfileName: _profile.ProfileName,
	})
	if err != nil {
		return err
	}

	return m.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {

		create_params := sqlc.CreateUserContestantProfileParams{
			UserID:           user,
			ProfileName:      _profile.ProfileName,
			DisplayName:      _profile.DisplayName,
			Bio:              utils.PgConvert.ToPgText(_profile.Bio),
			PhotoRoot:        utils.PgConvert.ToPgText(_profile.PhotoRoot),
			CoverPhoto:       utils.PgConvert.ToPgText(_profile.CoverPhoto),
			SocialLinks:      _profile.SocialLinks,
			WebsiteUrl:       utils.PgConvert.ToPgText(_profile.WebsiteUrl),
			Location:         utils.PgConvert.ToPgText(_profile.Location),
			OrganisationName: utils.PgConvert.ToPgText(_profile.OrganisationName),
			Category:         utils.PgConvert.ToPgText(_profile.Category),
			IsActive:         true,
			ProfileCompletedAt: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
		}
		ret, err := q.CreateUserContestantProfile(ctx, create_params)
		if err != nil {
			return err
		}
		if ret.UserID != create_params.UserID || ret.ProfileName != create_params.ProfileName {
			return errors.New("failed to create contestant profile")
		}
		return nil

	})
}

// Login authenticates a user using either email or phone, verifies the password,
// generates an encrypted token and optionally returns a filtered user profile.
func (m *Service) Login(ctx context.Context, key []byte, params LoginRequest) (*LoginResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Fetch user credentials based on the login mode
	var (
		user        sqlc.User
		err         error
		profile_ret json.RawMessage
	)

	switch params.Mode {
	case "email":
		user, err = m.Postgres.Query.GetUserByEmail(ctx, params.Identifier)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	case "phone":
		user, err = m.Postgres.Query.GetUserByPhone(ctx, utils.PgConvert.ToPgText(params.Identifier))
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	default:
		return nil, errors.New("invalid login mode")
	}

	// Verify password
	err = encrypt.CompareHash(params.Password, user.HashPassword)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := encrypt.Encrypt(&encrypt.Branca{
		Password: string(key),
		Expire:   30,
	},
		user.ID[:],
	)
	if err != nil {
		return nil, err
	}

	// Get user profile and filter fields if requested
	if params.ResultOpts != nil && len(params.ResultOpts.Schema) > 0 {
		// Fetch user profile
		profile := Profile{
			FirstName:        user.FirstName,
			MiddleName:       user.MiddleName.String,
			LastName:         user.LastName,
			DateOfBirth:      user.DateOfBirth.Time,
			Gender:           user.Gender,
			Country:          user.Country,
			Email:            user.Email,
			Phone:            user.Phone.String,
			PhoneVerified:    user.IsEmailVerified.Bool,
			EmailVerified:    user.IsEmailVerified.Bool,
			IdentityVerified: string(user.IsIdentityVerified),
		}

		// Filter profile based on requested schema
		filters, err := params.ResultOpts.FilterProfile(profile)
		if err != nil {
			return nil, err
		}

		profile_ret, err = json.Marshal(filters)
		if err != nil {
			return nil, err
		}

	}

	// 5. Return response
	return &LoginResponse{
		AuthorizationID: string(token),
		Profile:         &profile_ret,
	}, nil
}

func (m *Service) GetUserProfile(ctx context.Context, usr gmodels.User) (*Profile, error) {
	var (
		ret = new(Profile)
	)

	id, err := usr.ID(ctx, m.Postgres)
	if err != nil {
		return nil, err
	}
	sqlc_usr, err := m.Postgres.Query.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	err = ret.Filter(sqlc_usr)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// RequestDeactivation initiates user account deactivation by generating and sending an OTP verification code.
func (m *Service) RequestDeactivation(ctx context.Context, usr gmodels.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := usr.ID(ctx, m.Postgres)
	if err != nil {
		return err
	}

	user, err := m.Postgres.Query.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if user.Status == "deactivated" {
		return ErrUserAlreadyDeactivated
	}

	if err := m.Messenger.OTP.Create(ctx, id.String(), messenger.Email, nil, time.Time{}); err != nil {
		return err
	}

	return nil
}

// Deactivate completes user account deactivation after verifying the OTP code.
func (m *Service) Deactivate(ctx context.Context, usr gmodels.User, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := usr.ID(ctx, m.Postgres)
	if err != nil {
		return err
	}

	user, err := m.Postgres.Query.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if user.Status == "deactivated" {
		return ErrUserAlreadyDeactivated
	}

	if err := m.Messenger.OTP.Validate(ctx, id.String(), code); err != nil {
		return err
	}

	deactivatedAt := pgtype.Timestamptz{
		Time:  time.Now(),
		Valid: true,
	}

	_, err = m.Postgres.Query.UpdateUserSoftDelete(ctx, sqlc.UpdateUserSoftDeleteParams{
		ID:            id,
		Status:        "deactivated",
		DeactivatedAt: deactivatedAt,
	})
	if err != nil {
		return err
	}

	return nil
}

type RequestPassword struct {
	ID     uuid.UUID
	Medium string
	Expire time.Time
}

func (m *Service) RequestPasswordReset(ctx context.Context, usr gmodels.User) error {
	// id, err := usr.ID(ctx, m.Postgres)
	// if err != nil {
	// 	return err
	// }
	// err = messenger.OTP.Create(ctx, id.String(), messenger.Email, nil, time.Time{}, nil)

	return nil
}

func (m *Service) ResetPassword() {}
