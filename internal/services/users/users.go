package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	// "github.com/pat3icki/pennychoice/internal/states"

	"github.com/pat3icki/pennychoice/types"
)

func (s *Service) CreateUser(ctx context.Context, usr *CreateUserParams) (*CreateUserResult, error) {
	return nil, nil
}
func (s *Service) GetUserProfile(ctx context.Context, auth_id string) (*UserProfile, error)

func (s *Service) DeactiviateUser(ctx context.Context, usr types.User, expected_pass string, period time.Time)

func (s *Service) SuspendUser(ctx context.Context, usr_id uuid.UUID)

func (s *Service) Login(ctx context.Context, req_key int64, params LoginParams) (User, error)

func (s *Service) CreateRequestKey(ctx context.Context, req *RequestKey) error
