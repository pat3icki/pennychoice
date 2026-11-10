package gmodels

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/pkg/utils"
)

type Uniqueness string

const (
	UniquenessPhone    Uniqueness = "phone"
	UniquenessID                  = "id"
	UniquenessEmail    Uniqueness = "email"
	UniquenessUsername Uniqueness = "username"
)

type User struct {
	Unique Uniqueness
	Value  string
}

func (u User) IsValid() bool {
	return u.Unique != "" && u.Value != ""
}

func (u User) ID(ctx context.Context, db *db.Client) (uuid.UUID, error) {
	switch u.Unique {
	case UniquenessPhone:
		return db.Query.GetUserIDByPhone(ctx, utils.PgConvert.ToPgText(u.Value))
	case UniquenessID:
		return uuid.Parse(u.Value)
	case UniquenessEmail:
		return db.Query.GetUserIDByEmail(ctx, u.Value)

	default:
		return uuid.Nil, errors.New("invalid uniqueness")
	}
}
