package users

import (
	"context"
	"time"

	"github.com/pat3icki/pennychoice/types"
)

func (s *Service) CreateUser(ctx context.Context, usr *CreateUserParams) (*CreateUserResult, error) {
	s.mu.Lock()

	return nil, nil
}

func (s *Service) VerifiyFalse(ctx context.Context, usr types.User, ver types.FlagUniqueness) error {
	return nil
}

func (s *Service) GetUserProfile(ctx context.Context, usr types.User) (*UserProfile, error)

func (s *Service) DeactiviateUser(ctx context.Context, usr types.User, expected_pin string, period time.Time) {

}

// func (s *Service) Login(ctx context.Context, req_key int64, params LoginParams) (usr User, err error) {
// 	id_info := sflake.Describe(req_key, sflake.DefaultEpoch)
// 	if id_info.NodeID != REQUEST_KEY_SFLAKE_NODE {
// 		return User{}, errors.ErrUnsupported
// 	}
// 	_, err = s.Redis.Get(utils.Convert.IntBytes(req_key))
// 	if err != nil {
// 		return
// 	}
// 	switch params.User.Uniqueness {
// 	case types.Unique_Email:
// 		_, err := s.PostgreSQL.GetUserByEmail(ctx, params.User.Value)
// 		if err != nil {
// 			return User{}, err
// 		}
// 	case types.Unique_Phone:
// 		// TODO: database query
// 	default:
// 		err = errors.New("user uniqueness must be either phone or email")
// 		return
// 	}
// 	// TODO: complete
// 	return

// }
