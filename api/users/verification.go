package users

import (
	"context"

	"github.com/pat3icki/pennychoice/internal/services/users"
)

type UserVerification struct {
	AuthrizationID  string
	IsPhoneVerified bool
	IsEmailVerified bool
	IsNinVerified   bool
}

func GetUserVerification(ctx context.Context, service *users.Service, authID string)



func UpdateUserVerification(ctx context.Context, service *users.Service, authID string)