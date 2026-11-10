package organisation

import (
	"context"

	"github.com/google/uuid"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

// getMembers is a helper to fetch members for an organisation.
func (m *Service) getMembers(ctx context.Context, q sqlc.Querier, orgID uuid.UUID) ([]Member, error) {
	memberships, err := q.ListOrganisationMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	members := make([]Member, 0, len(memberships))
	for _, membership := range memberships {
		user, err := q.GetUserByID(ctx, membership.MemberID)
		if err != nil {
			return nil, err
		}
		members = append(members, Member{
			Name:     user.FirstName + " " + user.LastName,
			Email:    user.Email,
			Position: int(membership.Position),
			JoinedAt: membership.JoinedAt.Time,
			IsActive: user.Status == "active",
		})
	}
	return members, nil
}

// isMember checks if a user is a member of an organisation.
func (m *Service) isMember(ctx context.Context, q sqlc.Querier, orgID, userID uuid.UUID) (bool, error) {
	if q == nil {
		q = m.Postgres.Query
	}
	return q.ExistsOrganisationMember(ctx, sqlc.ExistsOrganisationMemberParams{
		OrgID:    orgID,
		MemberID: userID,
	})
}

