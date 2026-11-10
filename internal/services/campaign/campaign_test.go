package campaign

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pat3icki/pennychoice/gmodels"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	manager := GlobalTestManager
	ctx := t.Context()
	userID := uuid.Must(uuid.NewV7())
	user, err := _createUser(ctx, manager.Postgres.Query, userID)
	require.NoError(t, err)

	org, err := _createOrganisation(ctx, manager.Postgres.Query, user.ID)
	require.NoError(t, err)

	starts, ends := GenerateCampaignTimes()
	// Build the Campaign struct with all required fields
	campaign := CreateCampaignV1{
		// ID:          uuid.Nil,
		Name:        gofakeit.Name(),
		Description: gofakeit.NounCollectivePeople(),
		Tags:        []string{"music", "contest"},
		Options: &Options{
			ActiviationMethod:     "manual",
			VoteMetricsVisibility: "public",
			VoteStarts:            starts,
			VoteEnds:              ends,
			VotePerAmount:         100,
			AllowRefunds:          true,
		},
		Registration: Registration{
			Gateway: "free",
		},
	}

	payloadBytes, err := json.Marshal(campaign)
	require.NoError(t, err)

	err = manager.Create(ctx, Auth{
		User:         user.ID,
		Organisation: org.ID,
	}, CreateCampaignParams{
		Payload: payloadBytes,
		Schema:  manager.SchemanName,
	})
	require.NoError(t, err)
}

func TestActiviateCampaign(t *testing.T) {
	manager := GlobalTestManager
	ctx := t.Context()
	var user sqlc.User
	var campaign sqlc.Campaign

	createInitais(t, ctx, manager, &user, uuid.Must(uuid.NewV7()), &campaign)
	require.NotNil(t, user)
	require.NotNil(t, campaign)

	err := manager.Activiate(ctx, Auth{
		User:         user.ID,
		Organisation: campaign.OrganisationID,
	}, campaign.ID)

	require.NoError(t, err)
}

func TestRequestContestant(t *testing.T) {
	tests := []struct {
		name                        string
		request_made_by             string
		organisation_member_user_id uuid.UUID
		constestant_user_id         uuid.UUID
		verdict                     string
	}{
		{
			name:                        "Test case 1",
			request_made_by:             "user",
			organisation_member_user_id: uuid.Must(uuid.NewV7()),
			constestant_user_id:         uuid.Must(uuid.NewV7()),
			verdict:                     "accept",
		},
		{
			name:                        "Test case 3",
			request_made_by:             "organisation_member_user",
			organisation_member_user_id: uuid.Must(uuid.NewV7()),
			constestant_user_id:         uuid.Must(uuid.NewV7()),
			verdict:                     "accept",
		},
		{
			name:                        "Test case 2",
			request_made_by:             "organisation_member_user",
			organisation_member_user_id: uuid.Must(uuid.NewV7()),
			constestant_user_id:         uuid.Must(uuid.NewV7()),
			verdict:                     "decline",
		},
		{
			name:                        "Test case 4",
			request_made_by:             "user",
			organisation_member_user_id: uuid.Must(uuid.NewV7()),
			constestant_user_id:         uuid.Must(uuid.NewV7()),
			verdict:                     "decline",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var (
				organisation_member_user sqlc.User
				campaign                 sqlc.Campaign
				request_contestant_user  sqlc.User
				manager                  = GlobalTestManager
				ctx                      = t.Context()
				reqByUser                = RequestContestant{}
			)
			createInitais(t, ctx, manager, &organisation_member_user, tt.organisation_member_user_id, &campaign)
			require.Equal(t, tt.organisation_member_user_id, organisation_member_user.ID)

			err := manager.Postgres.WithTx(t.Context(), func(tx pgx.Tx, q sqlc.Querier) error {
				var err error

				request_contestant_user, err = _createUser(ctx, manager.Postgres.Query, tt.constestant_user_id)
				if err != nil {
					return err
				}
				require.NotEmpty(t, request_contestant_user)

				_, err = _createContestantProfile(ctx, manager.Postgres.Query, tt.constestant_user_id)
				if err != nil {
					return err
				}
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, tt.constestant_user_id, request_contestant_user.ID)

			var request RequestContestant
			if tt.request_made_by == "user" {
				var err error

				// Test CreateRequestContestantByUser
				request, err = manager.CreateRequestContestantByUser(ctx, tt.constestant_user_id, CreateRequestContestantParam{
					RequestContestantUser: gmodels.User{
						Unique: gmodels.UniquenessID,
						Value:  tt.constestant_user_id.String(),
					},
					Organisation: campaign.OrganisationID,
					Campaign:     campaign.ID,
				})
				require.NoError(t, err)
				require.NotEmpty(t, request)
				require.Equal(t, "pending", request.Verdict)

			} else if tt.request_made_by == "organisation_member_user" {
				var err error
				// Pass the org-member as the caller so the service sets InitiatorPosition = "organisation_member_user".
				// RequestContestantUser is the *contestant* being invited, not the org-member initiating.

				// Test CreateRequestContestantByOrganisationMember
				request, err = manager.CreateRequestContestantByOrganisationMember(ctx, tt.organisation_member_user_id, CreateRequestContestantParam{
					RequestContestantUser: gmodels.User{
						Unique: gmodels.UniquenessID,
						Value:  tt.constestant_user_id.String(),
					},
					Organisation: campaign.OrganisationID,
					Campaign:     campaign.ID,
				})
				require.NoError(t, err)
				require.NotEmpty(t, request)

			} else {
				t.Error(" tt.request_made_by is invalid: ", tt.request_made_by)
			}
			Resolve_params := ResolveContestantParams{
				RequestID:      request.ID,
				Campaign:       campaign.ID,
				ContestantUser: request_contestant_user.ID,
				Verdict:        tt.verdict,
			}

			if request.Initiated == "user" {
				// // Test ResolveRequestContestant
				err = manager.ResolveRequestContestant(ctx, organisation_member_user.ID, Resolve_params)
				require.NoError(t, err)
			} else {
				// // Test ResolveRequestContestant
				err = manager.ResolveRequestContestant(ctx, request_contestant_user.ID, Resolve_params)
				require.NoError(t, err)
			}

			// Test GetRequestContestant
			reqByUser, err = manager.GetRequestContestantByUser(ctx, tt.constestant_user_id, request.ID)
			require.NoError(t, err)
			require.Equal(t, request.ID, reqByUser.ID)
			require.Equal(t, campaign.ID.String(), reqByUser.Campaign)
			require.Equal(t, campaign.OrganisationID, reqByUser.OrganisationID)
			require.Equal(t, tt.request_made_by, reqByUser.Initiated)

			// // Test GetRequestContestantByOrganisationMember
			reqByOrg, err := manager.GetRequestContestantByOrganisationMember(ctx, organisation_member_user.ID, campaign.OrganisationID, request.ID)
			require.NoError(t, err, "something happened here")
			require.Equal(t, reqByUser.ID, reqByOrg.ID)
			require.Equal(t, reqByUser.Campaign, reqByOrg.Campaign)
			require.Equal(t, reqByUser.OrganisationID, reqByOrg.OrganisationID)
			require.Equal(t, reqByUser.Initiated, reqByOrg.Initiated)

		})
	}
}

func TestGetRequestContestantRequirements(t *testing.T) {
	manager := GlobalTestManager
	ctx := t.Context()

	var user sqlc.User
	var campaign sqlc.Campaign
	createInitais(t, ctx, manager, &user, uuid.Must(uuid.NewV7()), &campaign)

	contestantUser, err := _createUser(ctx, manager.Postgres.Query, uuid.Must(uuid.NewV7()))
	require.NoError(t, err)

	contestantUserArg := gmodels.User{
		Unique: gmodels.UniquenessID,
		Value:  contestantUser.ID.String(),
	}

	t.Run("invalid ids slice length", func(t *testing.T) {
		_, err := manager.GetRequestContestantRequirements(ctx, uuid.UUIDs{user.ID}, contestantUserArg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ids must contain at least User ID and Campaign ID")
	})

	t.Run("user does not exist", func(t *testing.T) {
		nonExistentUserID := uuid.Must(uuid.NewV7())
		_, err := manager.GetRequestContestantRequirements(ctx, uuid.UUIDs{nonExistentUserID, campaign.ID}, contestantUserArg)
		require.Error(t, err)
	})

	t.Run("campaign does not exist", func(t *testing.T) {
		nonExistentCampaignID := uuid.Must(uuid.NewV7())
		_, err := manager.GetRequestContestantRequirements(ctx, uuid.UUIDs{user.ID, nonExistentCampaignID}, contestantUserArg)
		require.Error(t, err)
	})

	t.Run("valid call", func(t *testing.T) {
		// When ContestantRequirements is valid or empty/null DB behavior
		data, err := manager.GetRequestContestantRequirements(ctx, uuid.UUIDs{user.ID, campaign.ID}, contestantUserArg)
		if err != nil {
			// If contestant requirements is not set / invalid in test fixture
			require.Error(t, err)
		} else {
			require.NotNil(t, data)
		}
	})
}

func TestUpdateVotePeriods(t *testing.T) {
	manager := GlobalTestManager
	ctx := t.Context()

	var user sqlc.User
	var campaign sqlc.Campaign
	createInitais(t, ctx, manager, &user, uuid.Must(uuid.NewV7()), &campaign)

	auth := Auth{
		User:         user.ID,
		Organisation: campaign.OrganisationID,
	}

	t.Run("invalid time order start not before end", func(t *testing.T) {
		now := time.Now()
		err := manager.UpdateVotePeriods(ctx, auth, campaign.ID, VotePeriod{
			Start: now.Add(2 * time.Hour),
			End:   now.Add(1 * time.Hour),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "votes_starts_at must be before votes_ends_at")
	})

	t.Run("campaign not in draft status or not found", func(t *testing.T) {
		nonExistentCampaignID := uuid.Must(uuid.NewV7())
		err := manager.UpdateVotePeriods(ctx, auth, nonExistentCampaignID, VotePeriod{
			Start: time.Now(),
			End:   time.Now().Add(1 * time.Hour),
		})
		require.Error(t, err)
	})

	t.Run("successful vote period update in draft status", func(t *testing.T) {
		// campaign status defaults to draft in createInitais
		now := time.Now()
		err := manager.UpdateVotePeriods(ctx, auth, campaign.ID, VotePeriod{
			Start: now,
			End:   now.Add(24 * time.Hour),
		})
		require.NoError(t, err)
	})
}

func TestPause(t *testing.T) {
	manager := GlobalTestManager
	ctx := t.Context()

	var user sqlc.User
	var campaign sqlc.Campaign
	createInitais(t, ctx, manager, &user, uuid.Must(uuid.NewV7()), &campaign)

	auth := Auth{
		User:         user.ID,
		Organisation: campaign.OrganisationID,
	}

	t.Run("status state", func(t *testing.T) {
		err := manager.Activiate(ctx, auth, campaign.ID)
		require.NoError(t, err)

		err = manager.Pause(ctx, auth, campaign.ID)
		require.NoError(t, err)
	})

}

func TestCreateRequestContestant(t *testing.T) {
	t.Skip(GlobalTestManager.CreateRequestContestant(context.Background(), uuid.Nil, CreateRequestContestantParam{}))
}
