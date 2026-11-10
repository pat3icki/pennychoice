package campaign

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	penneyErr "github.com/pat3icki/pennychoice/errors"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

func (m *Service) CreateRequestContestantByOrganisationMember(ctx context.Context, user uuid.UUID, args CreateRequestContestantParam) (RequestContestant, error) {
	var request_contestant RequestContestant
	//TODO: this is ineffeient having 5 sql round trips when it can be 3

	cUsrID, err := args.RequestContestantUser.ID(ctx, m.Postgres)
	if err != nil {
		return request_contestant, err
	}

	_, err = m.Postgres.Exists(args.Organisation, db.TableOrganisation)
	if err != nil {
		return request_contestant, fmt.Errorf("organisation `id` cannot be empty")
	}
	_, err = m.Postgres.Exists(args.Campaign, db.TableCampaigns)
	if err != nil {
		return request_contestant, fmt.Errorf("Campaign `id` cannot be empty")
	}

	err = m.Postgres.WithTx(ctx, func(_ pgx.Tx, q sqlc.Querier) error {
		campaign, err := q.GetCampaignByID(ctx, args.Campaign)
		if err != nil {
			return errors.New("campaign does not exists")
		}

		if campaign.OrganisationID != args.Organisation {
			return errors.New("campaign does not belong to the specified organisation")
		}
		is_contestant, err := m.Postgres.Query.ExistsCampaignContestant(ctx, sqlc.ExistsCampaignContestantParams{
			CampaignID: campaign.ID,
			UserID:     cUsrID,
		})
		if err != nil {
			return err // TODO: do error well
		}
		if is_contestant {
			return errors.New("do error well")
		}

		params := sqlc.CreateContestantJoinRequestParams{
			RequestContestantUserID: cUsrID,
			CampaignID:              args.Campaign,
			OrgID:                   args.Organisation,
			InitiatorPosition:       "organisation_member_user",
			InitiatedBy:             user,
		}
		ret, err := q.CreateContestantJoinRequest(ctx, params)
		if err != nil {
			return penneyErr.Database(PackageName, "CreateContestantJoinRequest", err)
		}
		if (ret.OrgID != params.OrgID) && (ret.CampaignID != params.CampaignID) {
			return fmt.Errorf("something is wrong with the database") // TODO: RETURN ERROR
		}
		request_contestant = RequestContestant{
			ID:             ret.RequestID,
			Campaign:       campaign.Name,
			OrganisationID: campaign.OrganisationID,
			Verdict:        ret.Verdict,
			Initiated:      ret.InitiatorPosition,
			Time:           ret.InvitedTimestamp.Time,
		}
		return nil
	})
	return request_contestant, err

}

func (m *Service) GetRequestContestantByOrganisationMember(ctx context.Context, usr uuid.UUID, org uuid.UUID, req int64) (RequestContestant, error) {
	_, err := m.Postgres.Exists(usr, db.TableUsers)
	if err != nil {
		return RequestContestant{}, penneyErr.Database(PackageName, "Exists", err)
	}
	_, err = m.Postgres.Exists(org, db.TableOrganisation)
	if err != nil {
		return RequestContestant{}, penneyErr.Database(PackageName, "Exists", err)
	}
	isMember, err := m.Postgres.Query.ExistsOrganisationMember(ctx, sqlc.ExistsOrganisationMemberParams{
		OrgID:    org,
		MemberID: usr,
	})
	if err != nil {
		return RequestContestant{}, penneyErr.Database(PackageName, "ExistsOrganisationMember", err)
	}
	if !isMember {
		return RequestContestant{}, errors.New("user is not an organisation member")
	}
	joinReq, err := m.Postgres.Query.GetContestantJoinRequestByID(ctx, req)
	if err != nil {
		return RequestContestant{}, penneyErr.Database(PackageName, "GetContestantJoinRequestByCampaignAnd", err)
	}
	return RequestContestant{
		ID:             joinReq.RequestID,
		Campaign:       joinReq.CampaignID.String(),
		OrganisationID: joinReq.OrgID,
		Verdict:        joinReq.Verdict,
		Initiated:      joinReq.InitiatorPosition,
		Time:           joinReq.InvitedTimestamp.Time,
	}, nil
}
