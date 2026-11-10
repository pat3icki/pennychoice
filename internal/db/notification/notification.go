package notification

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

type Encoder interface {
	Mode() string
	Encode() (sqlc.CreateNotificationParams, error)
	EncodeBatch() ([]sqlc.CreateNotificationParams, error)
	EncodeOrganisation()
}

func Create(ctx context.Context, tx pgx.Tx, encoder Encoder) error {
	return nil
}
