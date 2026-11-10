package campaign

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/pat3icki/pennychoice/pkg/utils"
	"github.com/stretchr/testify/require"
)

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

func _createContestantProfile(ctx context.Context, q sqlc.Querier, userID uuid.UUID) (sqlc.UserContestantProfile, error) {
	ret, err := q.CreateUserContestantProfile(ctx, sqlc.CreateUserContestantProfileParams{
		UserID:      userID,
		ProfileName: utils.RandomString(10),
		DisplayName: gofakeit.Name(),
		SocialLinks: []byte(`{}`),
		IsActive:    true,
	})
	if err != nil {
		return sqlc.UserContestantProfile{}, err
	}
	return ret, nil
}

func _createOrganisation(ctx context.Context, q sqlc.Querier, usr uuid.UUID) (sqlc.Organisation, error) {
	ret, err := q.CreateOrganisation(ctx, sqlc.CreateOrganisationParams{
		ID:            uuid.Must(uuid.NewV7()),
		Name:          utils.RandomString(10),
		CreatorUserID: usr,
	})
	if err != nil {
		return sqlc.Organisation{}, err
	}

	_, err = q.AddOrganisationMember(ctx, sqlc.AddOrganisationMemberParams{
		OrgID:    ret.ID,
		MemberID: usr,
		Position: 1,
	})
	if err != nil {
		return sqlc.Organisation{}, err
	}

	return ret, nil
}

func _createCampaign(ctx context.Context, id uuid.UUID, q sqlc.Querier, usr Auth) (sqlc.Campaign, error) {

	s, e := GenerateCampaignTimes()
	ret, err := q.CreateCampaign(ctx, sqlc.CreateCampaignParams{
		ID:                    id,
		OrganisationID:        usr.Organisation,
		CreatorUserID:         usr.User,
		Status:                "draft",
		BankNameSuffix:        "NDI",
		VotesStartsAt:         pgtype.Timestamptz{Time: s, Valid: true},
		VotesEndsAt:           pgtype.Timestamptz{Time: e, Valid: true},
		VoteMetricsVisibility: "private",
		PlatformFeePercentage: 20,
		ActivationMethod:      "schedule",
	})
	if err != nil {
		return ret, err
	}
	// require.Equal(t, ret.ID)
	return ret, nil
}

func createInitais(
	t *testing.T,
	ctx context.Context,
	manager *Service,
	user *sqlc.User,
	usr_id uuid.UUID,
	campaign *sqlc.Campaign) {
	if user == nil {
		user = new(sqlc.User)
	}
	if campaign == nil {
		campaign = new(sqlc.Campaign)
	}
	err := manager.Postgres.WithTx(t.Context(), func(tx pgx.Tx, q sqlc.Querier) error {
		var err error

		*user, err = _createUser(ctx, manager.Postgres.Query, usr_id)
		if err != nil {
			return err
		}
		require.NotEmpty(t, user)

		org, err := _createOrganisation(ctx, manager.Postgres.Query, user.ID)
		if err != nil {
			return err
		}
		require.NotEmpty(t, org)

		// Generate a new UUID for the campaign
		id, err := manager.Postgres.GenerateUUID(ctx, db.TableCampaigns)
		if err != nil {
			return err
		}

		*campaign, err = _createCampaign(ctx, id, manager.Postgres.Query, Auth{
			User:         usr_id,
			Organisation: org.ID,
		})
		if err != nil {
			return err
		}
		require.NotEmpty(t, campaign)
		return nil
	})
	require.NoError(t, err)
}

func _createNecessaryTableValues(t *testing.T, manager *Service) (user sqlc.User, campaign sqlc.Campaign, organisation sqlc.Organisation) {
	ctx := t.Context()
	err := manager.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {
		usr_id, err := uuid.NewV7()
		if err != nil {
			return err
		}

		user, err = _createUser(ctx, manager.Postgres.Query, usr_id)
		if err != nil {
			return err
		}
		require.NotEmpty(t, user)

		organisation, err = _createOrganisation(ctx, manager.Postgres.Query, user.ID)
		if err != nil {
			return err
		}
		require.NotEmpty(t, organisation)

		// Generate a new UUID for the campaign
		id, err := manager.Postgres.GenerateUUID(ctx, db.TableCampaigns)
		if err != nil {
			return err
		}

		campaign, err = _createCampaign(ctx, id, manager.Postgres.Query, Auth{
			User:         usr_id,
			Organisation: organisation.ID,
		})
		if err != nil {
			return err
		}
		require.NotEmpty(t, campaign)
		return nil
	})
	require.NoError(t, err)
	return
}

func _createUserConstestant(t *testing.T, manager *Service) (user sqlc.User) {
	ctx := t.Context()
	err := manager.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {
		usr_id, err := uuid.NewV7()
		if err != nil {
			return err
		}

		user, err = _createUser(ctx, manager.Postgres.Query, usr_id)
		if err != nil {
			return err
		}
		require.NotEmpty(t, user)
		return nil
	})
	require.NoError(t, err)
	return
}
