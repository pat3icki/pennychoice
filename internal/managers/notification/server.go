package notification

import (
	"context"

	"github.com/google/uuid"
)

type Server struct {
}

type CreateUserNotification struct {
	ID          uuid.UUID
	Title       string
	Description string
	Source      Source
}

type Source interface {
	Out() (string, uuid.UUID)
}

type CreateOrgansiationNotification struct {
	ID          uuid.UUID
	Title       string
	Description string
}

type Impl interface {
	CreateUser(ctx context.Context, usr CreateUserNotification) error
	CreateOrgansiation(ctx context.Context, org CreateOrgansiationNotification) error
}
