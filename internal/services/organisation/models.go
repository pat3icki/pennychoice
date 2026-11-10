package organisation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ServiceImpl interface {
	Create(ctx context.Context, user_id uuid.UUID, args CreateParams) (*Organisation, error)
	Get(ctx context.Context, user_id uuid.UUID, org_id uuid.UUID) (Organisation, error)
	List(ctx context.Context, userID uuid.UUID) ([]Organisation, error)

	GetPermission(ctx context.Context, orgID uuid.UUID) ([]byte, error)
	UpdatePermission(ctx context.Context, orgID uuid.UUID, args map[string]string) ([]byte, error)
}

type CreateParams struct {
	Name        string
	Description string
	Permissions json.RawMessage
}

type Organisation struct {
	ID                uuid.UUID
	Name              string
	Description       string
	Balance           string
	Permissions       json.RawMessage
	TotalCampaign     uint64
	TotalMembers      uint64
	ActiveCampaign    int16 `db:"active_campaign" json:"active_campaign"`
	MaxCoOrganisers   int16 `db:"max_co_organisers" json:"max_co_organisers"`
	MaxActiveCampaign int16 `db:"max_active_campaign" json:"max_active_campaign"`
}

type Members []Member

type Member struct {
	Name     string
	Email    string
	Position int
	JoinedAt time.Time
	IsActive bool
}

type Auth struct {
	UserID         uuid.UUID
	OrganisationID uuid.UUID
}

type CreateDecisionMakingParams struct {
	Organisation uuid.UUID
	ActionType   string
	Message      string
}
