package campaign

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	penneyErr "github.com/pat3icki/pennychoice/errors"
	"github.com/pat3icki/pennychoice/gmodels"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

func (m *Service) GetRequestContestantByUser(ctx context.Context, usr uuid.UUID, req int64) (RequestContestant, error) {
	_, err := m.Postgres.Exists(usr, db.TableUsers)
	if err != nil {
		return RequestContestant{}, penneyErr.Database(PackageName, "Exists", err)
	}

	joinReq, err := m.Postgres.Query.GetContestantJoinRequestForUser(ctx, sqlc.GetContestantJoinRequestForUserParams{
		RequestID:               req,
		RequestContestantUserID: usr,
	})
	if err != nil {
		return RequestContestant{}, penneyErr.Database(PackageName, "GetContestantJoinRequestForUser", err)
	}
	org, err := m.Postgres.Query.GetOrganisationByID(ctx, joinReq.OrgID)
	if err != nil {
		return RequestContestant{}, penneyErr.Database(PackageName, "GetOrganisationByID", err)
	}
	return RequestContestant{
		ID:             joinReq.RequestID,
		Campaign:       joinReq.CampaignID.String(),
		OrganisationID: org.ID,
		Verdict:        joinReq.Verdict,
		Initiated:      joinReq.InitiatorPosition,
		Time:           joinReq.InvitedTimestamp.Time,
	}, nil

}

// Decreated
func (m *Service) CreateRequestContestant(ctx context.Context, user uuid.UUID, args CreateRequestContestantParam) (RequestContestant, error) {
	var request_contestant RequestContestant

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

		omem, err := q.GetOrganisationMember(ctx, sqlc.GetOrganisationMemberParams{
			OrgID:    args.Organisation,
			MemberID: user,
		})

		var params sqlc.CreateContestantJoinRequestParams

		if err == nil && user == omem.MemberID {
			// Organisation Member User initiating the Request
			params = sqlc.CreateContestantJoinRequestParams{
				RequestContestantUserID: cUsrID,
				CampaignID:              args.Campaign,
				OrgID:                   omem.OrgID,
				InitiatorPosition:       "organisation_member_user",
				InitiatedBy:             omem.MemberID,
			}
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			// Unexpected DB error — propagate it
			return penneyErr.Database(PackageName, "GetOrganisationMember", err)
		} else {
			// Plain user self-requesting to join the campaign
			params = sqlc.CreateContestantJoinRequestParams{
				RequestContestantUserID: cUsrID,
				CampaignID:              args.Campaign,
				OrgID:                   args.Organisation,
				InitiatorPosition:       "user",
				InitiatedBy:             user,
			}
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
		// TODO: Code Notification
		// TODO: Code Auditing
		return nil
	})
	return request_contestant, err
}

func (m *Service) CreateRequestContestantByUser(ctx context.Context, user uuid.UUID, args CreateRequestContestantParam) (RequestContestant, error) {
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
			InitiatorPosition:       "user",
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

func (m *Service) ResolveRequestContestant(ctx context.Context, user uuid.UUID, args ResolveContestantParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if args.Verdict != "accept" && args.Verdict != "decline" {
		return errors.New("invalid verdict: must be 'accept' or 'decline'")
	}
	campaign, err := m.Postgres.Query.GetCampaignByID(ctx, args.Campaign)
	if err != nil {
		return penneyErr.Database(PackageName, "GetCampaignByID", err)
	}

	if campaign.Status != "draft" {
		return errors.New("campaign must be in draft status to resolve a contestant request")
	}
	userReq, err := _userRequestExists(ctx, m.Postgres.Query, args.RequestID)
	if err != nil {
		return err
	}
	exists, err := m.Postgres.Query.GetContestantByUserAndCampaign(ctx, sqlc.GetContestantByUserAndCampaignParams{
		CampaignID: campaign.ID,
		UserID:     args.ContestantUser,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			goto Resolve
		}
		return penneyErr.Database(PackageName, "GetContestantByUserAndCampaign", err)
	}

	if exists.UserID == args.ContestantUser {
		return errors.New("contestant already exists in this campaign")
	}

Resolve:
	var ValueParams = _resolveContestantParams{
		Request:  userReq,
		Campaign: campaign,
		Verdict:  args.Verdict,
	}
	if args.Verdict == "accept" {
		// TODO: If sql.ErrNoRow just get the user names and add it together to display name
		user_p, err := m.Postgres.Query.GetUserContestantProfileByUserID(ctx, args.ContestantUser)
		if err != nil {
			return penneyErr.Database(PackageName, "GetUserContestantProfileByUserID", err)
		}
		ValueParams.UserContestantProfile = user_p
	}
	var ret_err error

	// If an organisation member sends a request/invite to be a constestant then
	// its the user that has the power to verdict the Request

	is_user_org_member, err := m.Postgres.Query.ExistsOrganisationMember(ctx, sqlc.ExistsOrganisationMemberParams{
		OrgID:    campaign.OrganisationID,
		MemberID: user,
	})
	if err != nil {
		return penneyErr.Database(PackageName, "ExistsOrganisationMember", err)
	}
	// Resolve based on initiator position and user permissions
	switch userReq.InitiatorPosition {
	case "organisation_member_user":
		if !is_user_org_member {
			ret_err = m._user_ResolveContestant(ctx, user, ValueParams)
		} else {
			return errors.New("user does not have the permission to verdict the request")
		}
	case "user":
		if is_user_org_member {
			ret_err = m._orgMember_ResolveContestant(ctx, user, ValueParams)
		} else {
			return errors.New("user does not have the permission to verdict the request")
		}
	default:
		return fmt.Errorf("unknown initiator position: %s", userReq.InitiatorPosition)
	}
	if ret_err != nil {
		return ret_err
	}
	// TODO: Code Notification
	// TODO: Code Auditing
	return nil
}

// GetRequestContestantRequirements gets the user request contestant requirements.
// ids - 1 - User ID
// ids - 2 - Campaign ID
func (m *Service) GetRequestContestantRequirements(ctx context.Context, ids uuid.UUIDs, contestant gmodels.User) ([]byte, error) {
	if len(ids) < 2 {
		return nil, errors.New("ids must contain at least User ID and Campaign ID")
	}
	userID := ids[0]
	campaignID := ids[1]

	cUsrID, err := contestant.ID(ctx, m.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to get contestant user ID: %w", err)
	}

	_, err = m.Postgres.Exists(userID, db.TableUsers)
	if err != nil {
		return nil, penneyErr.Database(PackageName, "Exists User", err)
	}

	campaign, err := m.Postgres.Query.GetCampaignByID(ctx, campaignID)
	if err != nil {
		return nil, penneyErr.Database(PackageName, "GetCampaignByID", err)
	}
	if !campaign.ContestantRequirements.Valid {
		return nil, errors.New("")
	}
	return m.Postgres.Query.GetContestantRequirementData(ctx, cUsrID)

}
