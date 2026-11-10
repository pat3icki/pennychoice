package organisation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pat3icki/pennychoice/gmodels"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/pat3icki/pennychoice/pkg/utils"
)

type Service struct {
	Postgres *db.Client
}

// Create creates a new organisation with the given user as the owner/admin.
func (m *Service) Create(ctx context.Context, user_id uuid.UUID, args CreateParams) (Organisation, Members, error) {
	id, err := m.Postgres.GenerateUUID(ctx, db.TableOrganisation)
	if err != nil {
		return Organisation{}, Members{}, err
	}

	var (
		retOrg  Organisation
		members Members
	)

	err = m.Postgres.WithTx(ctx, func(tx pgx.Tx, q sqlc.Querier) error {
		org, err := q.CreateOrganisation(ctx, sqlc.CreateOrganisationParams{
			ID:            id,
			Name:          args.Name,
			Description:   utils.PgConvert.ToPgText(args.Description),
			CreatorUserID: user_id,
			Permissions:   args.Permissions,
		})
		if err != nil {
			return err
		}

		balance := money.New(org.Balance, money.NGN)

		added, err := q.AddOrganisationMember(ctx, sqlc.AddOrganisationMemberParams{
			OrgID:    org.ID,
			MemberID: user_id,
			Position: 1, // owner/admin
			InvitedBy: pgtype.UUID{
				Bytes: user_id,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}

		user, err := q.GetUserByID(ctx, user_id)
		if err != nil {
			return err
		}

		name := user.FirstName + " " + user.LastName

		retOrg = Organisation{
			ID:            org.ID,
			Name:          org.Name,
			Description:   org.Description.String,
			Balance:       balance.Display(),
			Permissions:   org.Permissions,
			TotalCampaign: uint64(org.TotalCampaign),
			TotalMembers:  1,
		}
		members =
			[]Member{
				{
					Name:     name,
					Email:    user.Email,
					Position: int(added.Position),
					JoinedAt: added.JoinedAt.Time,
					IsActive: user.Status == "active",
				},
			}

		return nil
	})

	if err != nil {
		return Organisation{}, Members{}, err
	}
	return retOrg, members, nil
}

// Get retrieves an organisation by its ID and includes its members.
// The parameter name is user_id to match the original stub, but it is actually the organisation ID.
func (m *Service) Get(ctx context.Context, user_id uuid.UUID, org_id uuid.UUID) (Organisation, error) {
	q := m.Postgres.Query

	_, err := m.Postgres.Exists(user_id, db.TableUsers)
	if err != nil {
		return Organisation{}, err
	}
	_, err = m.Postgres.Exists(org_id, db.TableOrganisation)
	if err != nil {
		return Organisation{}, err
	}

	db_ret, err := q.GetOrganisationByMember(ctx, sqlc.GetOrganisationByMemberParams{
		OrgID:    org_id,
		MemberID: user_id,
	})
	if err != nil {
		return Organisation{}, err
	}

	balance := money.New(db_ret.Balance, money.NGN)

	return Organisation{
		ID:            db_ret.ID,
		Name:          db_ret.Name,
		Description:   db_ret.Description.String,
		Balance:       balance.Display(),
		Permissions:   db_ret.Permissions,
		TotalCampaign: uint64(db_ret.TotalCampaign),
		TotalMembers:  uint64(db_ret.TotalMembers),
	}, nil
}

// List returns all organisations that the given user is a member of.
func (m *Service) List(ctx context.Context, userID uuid.UUID) ([]Organisation, error) {
	q := m.Postgres.Query

	list, err := q.ListOrganisationByUser(ctx, sqlc.ListOrganisationByUserParams{
		MemberID: userID,
		Limit:    10,
	})
	if err != nil {
		return nil, err
	}
	orgs := make([]Organisation, len(list))
	for _, org := range list {
		orgs = append(orgs, Organisation{
			ID:          org.ID,
			Name:        org.Name,
			Description: org.Description.String,
			Balance:     money.New(org.Balance, money.NGN).Display(),
			Permissions: org.Permissions,
		})
	}

	return orgs, nil
}

// GetPermission gets the organisation permissions for the given organisation ID.
func (m *Service) GetPermission(ctx context.Context, orgID uuid.UUID) ([]byte, error) {
	org, err := m.Postgres.Query.GetOrganisationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return org.Permissions, nil
}

// UpdatePermission sets the organisation permissions keys.
// It gets the organisation permission from the database, updates the specified key-values, and saves back to the database.
func (m *Service) UpdatePermission(ctx context.Context, orgID uuid.UUID, args map[string]string) ([]byte, error) {
	org, err := m.Postgres.Query.GetOrganisationByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	currentPerms := make(map[string]any)
	if len(org.Permissions) > 0 {
		if err := json.Unmarshal(org.Permissions, &currentPerms); err != nil {
			return nil, fmt.Errorf("failed to unmarshal current permissions: %w", err)
		}
	}

	for k, v := range args {
		currentPerms[k] = v
	}
	updatedBytes, err := json.Marshal(currentPerms)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated permissions: %w", err)
	}

	err = m.Postgres.Query.UpdateOrganisationPermissions(ctx, sqlc.UpdateOrganisationPermissionsParams{
		ID:          org.ID,
		Permissions: updatedBytes,
	})
	if err != nil {
		return nil, err
	}
	return updatedBytes, nil
}

// CreateDecisionMaking creates an organisation decision record.
func (m *Service) CreateDecisionMaking(ctx context.Context, usr uuid.UUID, args CreateDecisionMakingParams) error {

	_, err := m.Postgres.Query.CreateOrganisationDecision(ctx, sqlc.CreateOrganisationDecisionParams{
		// ID:               id,
		OrganisationID:   args.Organisation,
		ActionType:       args.ActionType,
		Message:          args.Message,
		AdminInitiatorID: usr,
	})
	return err
}

// Verdict determines and updates the verdict on the decision.
func (m *Service) Verdict(ctx context.Context, auth Auth, decision int64, verdict string) error {
	_, err := m.Postgres.Query.GetOrganisationMember(ctx, sqlc.GetOrganisationMemberParams{
		OrgID:    auth.OrganisationID,
		MemberID: auth.UserID,
	})
	if err != nil {
		return errors.New("unauthorized: user is not an organisation member")
	}

	_, err = m.Postgres.Query.UpdateOrganisationDecisionStatus(ctx, sqlc.UpdateOrganisationDecisionStatusParams{
		ID:     decision,
		Status: verdict,
	})
	return err
}

// Deactivate is a placeholder; the current schema has no "active" flag for organisations.
// We return an error indicating that deactivation is not supported.
func (m *Service) Deactivate(ctx context.Context, orgID uuid.UUID) error {
	return errors.New("deactivation of organisations is not supported in the current schema")
}

func (m *Service) CreateRequestAddMemberByUser(ctx context.Context, user uuid.UUID, id uuid.UUID) error {
	return nil
}
func (m *Service) CreateRequestAddMemberByOrgMember(ctx context.Context, auth Auth, request_mem gmodels.User) error {
	return nil
}

func (m *Service) ResolveRequestAddMember() {}

func (m *Service) GetMembers(ctx context.Context, auth Auth) (Members, error) {
	isMem, err := m.isMember(ctx, nil, auth.OrganisationID, auth.UserID)
	if err != nil {
		return nil, err
	}
	if !isMem {
		return nil, errors.New("unauthorized: user is not an organisation member")
	}
	return m.getMembers(ctx, m.Postgres.Query, auth.OrganisationID)
}
