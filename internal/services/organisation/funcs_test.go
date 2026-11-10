package organisation

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/stretchr/testify/require"
)

func createInitais(
	t *testing.T,
	ctx context.Context,
	manager *Service,
	user *sqlc.User) {
	err := manager.Postgres.WithTx(t.Context(), func(tx pgx.Tx, q sqlc.Querier) error {
		usr_id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		*user, err = _createUser(ctx, q, usr_id)
		if err != nil {
			return err
		}
		require.NotEmpty(t, user)
		return nil
	})

	require.NoError(t, err)
}

func _createUser(ctx context.Context, q sqlc.Querier, id uuid.UUID) (sqlc.User, error) {
	ret, err := q.CreateUserProfileFULL(ctx, sqlc.CreateUserProfileFULLParams{
		ID:           id,
		FirstName:    gofakeit.FirstName(),
		MiddleName:   pgtype.Text{String: gofakeit.MiddleName(), Valid: true},
		LastName:     gofakeit.LastName(),
		DateOfBirth:  pgtype.Date{Time: gofakeit.Date(), Valid: true},
		Gender:       "M",
		Status:       "active",
		Email:        gofakeit.Email(),
		Country:      "country",
		HashPassword: []byte(""),
	})
	if err != nil {
		return sqlc.User{}, err
	}
	return ret, nil
}

func TestIsMember(t *testing.T) {
	ctx := t.Context()
	manager := GlobalTestManager
	var user sqlc.User
	createInitais(t, ctx, manager, &user)

	org, _, err := manager.Create(ctx, user.ID, CreateParams{
		Name:        gofakeit.Name(),
		Description: gofakeit.NounAbstract(),
	})
	require.NoError(t, err)

	isMem, err := manager.isMember(ctx, nil, org.ID, user.ID)
	require.NoError(t, err)
	require.True(t, isMem)

	var nonMemberUser sqlc.User
	createInitais(t, ctx, manager, &nonMemberUser)
	isMem, err = manager.isMember(ctx, nil, org.ID, nonMemberUser.ID)
	require.NoError(t, err)
	require.False(t, isMem)
}
