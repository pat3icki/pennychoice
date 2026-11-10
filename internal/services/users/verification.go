package users

import (
	"context"

	"github.com/google/uuid"
)

func (m *Service) InitiateVerification(ctx context.Context, user uuid.UUID, mode string) error {

	return nil
}

func (m *Service) Verify(ctx context.Context, user uuid.UUID, code string) error {
	if err := m.Messenger.OTP.Validate(ctx, user.String(), code); err != nil {
		return err
	}
	return nil
}

// func (m *Service) ValidateIdentity()
