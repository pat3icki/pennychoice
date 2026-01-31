package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pat3icki/pennychoice/types"
)

type Service struct {
}

func (s Service) CreateUser(ctx context.Context, usr *CreateUserParams) (*CreateUserResult, error)

func (s *Service) GetUserProfile(ctx context.Context, auth_id string) (*UserProfile, error)

func (s *Service) DeactiviateUser(ctx context.Context, usr types.User, expected_pass string, period time.Time)

func (s *Service) SuspendUser(ctx context.Context, usr_id uuid.UUID)
