package campaign

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	penneyErr "github.com/pat3icki/pennychoice/errors"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

type _resolveContestantParams struct {
	Request               sqlc.RequestCampaignContestant
	Campaign              sqlc.Campaign
	UserContestantProfile sqlc.UserContestantProfile
	Verdict               string
}

func (m *Service) _user_ResolveContestant(ctx context.Context, user uuid.UUID, req _resolveContestantParams) error {
	// validation check
	if user != req.Request.RequestContestantUserID {
		return errors.New("unauthorized: user is not the intended contestant for this request")
	}
	return m.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {
		var err error

		_, err = q.ResolveContestantJoinRequest(ctx, sqlc.ResolveContestantJoinRequestParams{
			RequestID:        req.Request.RequestID,
			Verdict:          req.Verdict,
			VerdictTimestamp: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		})
		if err != nil {
			return penneyErr.Database(PackageName, "ResolveContestantJoinRequest", err)
		}

		if req.Verdict == "accept" {
			contestantID, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("generate contestant uuid: %w", err)
			}
			_, err = q.AddContestant(ctx, sqlc.AddContestantParams{
				ID:          contestantID,
				CampaignID:  req.Campaign.ID,
				UserID:      req.Request.RequestContestantUserID,
				DisplayName: req.UserContestantProfile.DisplayName,
				PhotoRoot:   req.UserContestantProfile.PhotoRoot,
				SocialLinks: req.UserContestantProfile.SocialLinks,
			})
			if err != nil {
				return penneyErr.Database(PackageName, "AddContestant", err)
			}

			err = q.UpdateCampaignContestantCount(ctx, sqlc.UpdateCampaignContestantCountParams{
				ID:               req.Campaign.ID,
				TotalContestants: req.Campaign.TotalContestants + 1,
			})
			if err != nil {
				return penneyErr.Database(PackageName, "UpdateCampaignContestantCount", err)
			}
		}
		return nil
	})

}

func (m *Service) _orgMember_ResolveContestant(ctx context.Context, user uuid.UUID, req _resolveContestantParams) error {
	// TODO: DONT FORGET THE VERDICT CONTROL FLOW

	exists, err := m.Postgres.Query.ExistsOrganisationMember(ctx, sqlc.ExistsOrganisationMemberParams{
		OrgID:    req.Campaign.OrganisationID,
		MemberID: user,
	})
	if err != nil {
		return penneyErr.Database(PackageName, "ExistsOrganisationMember", err)
	}
	if !exists {
		return errors.New("user not a member of the org")
	}

	return m.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {
		_, err := q.ResolveContestantJoinRequest(ctx, sqlc.ResolveContestantJoinRequestParams{
			RequestID:        req.Request.RequestID,
			Verdict:          req.Verdict,
			VerdictTimestamp: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		})
		if err != nil {
			return penneyErr.Database(PackageName, "ResolveContestantJoinRequest", err)
		}
		if req.Verdict == "accept" {
			contestantID, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("generate contestant uuid: %w", err)
			}
			_, err = q.AddContestant(ctx, sqlc.AddContestantParams{
				ID:          contestantID,
				CampaignID:  req.Campaign.ID,
				UserID:      req.Request.RequestContestantUserID,
				DisplayName: req.UserContestantProfile.DisplayName,
				PhotoRoot:   req.UserContestantProfile.PhotoRoot,
				SocialLinks: req.UserContestantProfile.SocialLinks,
			})
			if err != nil {
				return penneyErr.Database(PackageName, "AddContestant", err)
			}
			err = q.UpdateCampaignContestantCount(ctx, sqlc.UpdateCampaignContestantCountParams{
				ID:               req.Campaign.ID,
				TotalContestants: req.Campaign.TotalContestants + 1,
			})
			if err != nil {
				return penneyErr.Database(PackageName, "UpdateCampaignContestantCount", err)
			}
		}
		return err
	})
}

func _userRequestExists(ctx context.Context, q sqlc.Querier, req int64) (sqlc.RequestCampaignContestant, error) {
	request, err := q.GetContestantJoinRequestByID(ctx, req)
	if errors.Is(err, sql.ErrNoRows) {
		return sqlc.RequestCampaignContestant{}, err // This user hasnt request campaign constentent
	}
	if err != nil {
		return sqlc.RequestCampaignContestant{}, penneyErr.Database(PackageName, "GetContestantJoinRequestByCampaignAnd", err)
	}
	return request, nil
}
