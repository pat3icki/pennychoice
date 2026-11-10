package campaign

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	perr "github.com/pat3icki/pennychoice/errors"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/internal/db/notification"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/pat3icki/pennychoice/pkg/jschema"
)

func (m *Service) Create(ctx context.Context, auth Auth, params CreateCampaignParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		_campaign        Campaign
		id               uuid.UUID
		registration_fee int64
	)
	// Validate payload against JSON schema if schema is provided or registered
	if params.Schema != "" {
		err := jschema.Validate(params.Schema, params.Payload, nil)
		if err != nil {
			return fmt.Errorf("json schema: %w", err)
		}
	} else {
		return errors.New("parameter 'schema' can not be empty")
	}
	// Unmarshal payload into struct
	err := json.Unmarshal(params.Payload, &_campaign)
	if err != nil {
		return err
	}
	// Generate a new UUID for the campaign
	id, err = m.Postgres.GenerateUUID(ctx, db.TableCampaigns)
	if err != nil {
		return perr.Database(PackageName, "GenerateUUID", err)
	}

	if _campaign.Registration.Gateway == "free" {
		registration_fee = 0
	}

	return m.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {
		var (
			voteStarts            pgtype.Timestamptz
			voteEnds              pgtype.Timestamptz
			voteMetricsVisibility = "private"
			votePerAmount         int32
			allowRefunds          bool
		)

		if _campaign.Options != nil {
			if !_campaign.Options.VoteStarts.IsZero() {
				voteStarts = pgtype.Timestamptz{Time: _campaign.Options.VoteStarts, Valid: true}
			}
			if !_campaign.Options.VoteEnds.IsZero() {
				voteEnds = pgtype.Timestamptz{Time: _campaign.Options.VoteEnds, Valid: true}
			}
			if _campaign.Options.VoteMetricsVisibility != "" {
				voteMetricsVisibility = _campaign.Options.VoteMetricsVisibility
			}
			votePerAmount = int32(_campaign.Options.VotePerAmount)
			allowRefunds = _campaign.Options.AllowRefunds
		}

		tagsJSON, _ := json.Marshal(_campaign.Tags)
		creatParams := sqlc.CreateCampaignParams{
			ID:             id,
			OrganisationID: auth.Organisation,
			CreatorUserID:  auth.User,
			Name:           _campaign.Name,
			Description:    pgtype.Text{String: _campaign.Description, Valid: _campaign.Description != ""},
			Status:         "draft",
			BankNameSuffix: "NDI",
			Tags:           pgtype.Text{String: string(tagsJSON), Valid: len(_campaign.Tags) > 0},
			// VoteCurrency:     "NGN",
			VotePerOneAmount:      votePerAmount,
			ActivationMethod:      _campaign.Options.ActiviationMethod,
			VoteMetricsVisibility: voteMetricsVisibility,
			AllowRefunds:          allowRefunds,
			VotesStartsAt:         voteStarts,
			VotesEndsAt:           voteEnds,
			RegistrationFee:       registration_fee,
			PlatformFeePercentage: 20,
		}

		_, err := q.CreateCampaign(ctx, creatParams)
		if err != nil {
			return perr.Database(PackageName, "CreateCampaign", err)
		}
		err = q.AddTotalCampaign(ctx, auth.Organisation)
		if err != nil {
			return perr.Database(PackageName, "AddTotalCampaign", err)
		}
		return nil
	})
}

// func (m *Service) Filter(ctx context.Context, auth Auth, filters []string)

func (m *Service) List(ctx context.Context, auth Auth) ([]Campaign, error) {
	return nil, nil
}

func (m *Service) Activiate(ctx context.Context, auth Auth, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	campaign, err := m.Postgres.Query.GetCampaignByID(ctx, id)
	if err != nil {
		return perr.Database(PackageName, "GetCampaignByID", err)
	}
	if campaign.Status != "draft" {
		return errors.New("campaign must be in draft status")
	}
	_, err = m.Postgres.Query.GetOrganisationMember(ctx, sqlc.GetOrganisationMemberParams{
		OrgID:    campaign.OrganisationID,
		MemberID: auth.User,
	})
	if err != nil {
		return perr.Database(PackageName, "GetOrganisationMember", err)
	}

	return m.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {
		uid, err := q.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
			ID:     id,
			Status: "active",
		})
		if err != nil {
			return perr.Database(PackageName, "UpdateCampaignStatus", err)
		}
		if id != uid {
			return errors.ErrUnsupported // TODO: Correct Error
		}
		err = q.AddActiveCampaignInOrganisation(ctx, campaign.OrganisationID)
		if err != nil {
			return perr.Database(PackageName, "AddActiveCampaignInOrganisation", err)
		}
		// TODO: Code Notification
		err = notification.Create(ctx, tx, nil)
		if err != nil {
			return err
		}
		// TODO: Code Auditing
		return nil
	})
}

func (m *Service) Pause(ctx context.Context, auth Auth, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	campaign, err := m.Postgres.Query.GetCampaignByID(ctx, id)
	if err != nil {
		return perr.Database(PackageName, "GetCampaignByID", err)
	}
	if campaign.Status != "active" {
		return errors.New("campaign must be in active status")
	}
	_, err = m.Postgres.Query.GetOrganisationMember(ctx, sqlc.GetOrganisationMemberParams{
		OrgID:    campaign.OrganisationID,
		MemberID: auth.User,
	})
	if err != nil {
		return perr.Database(PackageName, "GetOrganisationMember", err)
	}

	return m.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {
		uid, err := q.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
			ID:     id,
			Status: "paused",
		})
		if err != nil {
			return perr.Database(PackageName, "UpdateCampaignStatus", err)
		}
		if id != uid {
			return errors.ErrUnsupported // TODO: Correct Error
		}
		return nil
	})
}

func (m *Service) UpdateVotePeriods(ctx context.Context, auth Auth, id uuid.UUID, args VotePeriod) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	q := m.Postgres.Query

	campaign, err := q.ExistCampaignReturn(ctx, sqlc.ExistCampaignReturnParams{
		MemberID:       auth.User,
		OrganisationID: auth.Organisation,
		ID:             id,
	})
	if campaign.Status != "draft" {
		return errors.New("campaign must be in draft status")
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No matching campaign found – return a 404 or similar

			return errors.New("Campaign not found or access denied")
		}
		// Unexpected error – log and return 500
		return errors.New("Campaign not found or access denied")
	}
	return m.Postgres.WithTx(ctx, func(_ pgx.Tx, q sqlc.Querier) error {
		if !args.Start.Before(args.End) {
			return fmt.Errorf("votes_starts_at must be before votes_ends_at")
		}
		err := q.AdjustVoteDuration(ctx, sqlc.AdjustVoteDurationParams{
			VotesStartsAt: pgtype.Timestamptz{
				Time:  args.Start,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}
		return nil
	})
}
