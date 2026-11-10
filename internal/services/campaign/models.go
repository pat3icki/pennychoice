package campaign

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pat3icki/pennychoice/gmodels"
	"github.com/pat3icki/pennychoice/internal/db"
)

type Service struct {
	mu          sync.Mutex
	Postgres    *db.Client
	SchemanName string
}

type CampaignServiceImpl interface {
	Create(ctx context.Context, auth Auth, params CreateCampaignParams) error
	Activiate(ctx context.Context, auth Auth, id uuid.UUID) error
	UpdateVotePeriods(ctx context.Context, auth Auth, id uuid.UUID, args VotePeriod) error
	Pause(ctx context.Context, auth Auth, id uuid.UUID) error
	GetRequestContestantByUser(ctx context.Context, usr uuid.UUID, req int64) (RequestContestant, error)
	GetRequestContestantByOrganisation(ctx context.Context, usr uuid.UUID, org uuid.UUID, req int64) (RequestContestant, error)
	CreateRequestContestant(ctx context.Context, user uuid.UUID, args CreateRequestContestantParam) (RequestContestant, error)
	ResolveRequestContestant(ctx context.Context, user uuid.UUID, args ResolveContestantParams) error
	GetRequestContestantRequirements(ctx context.Context, ids uuid.UUIDs, contestant gmodels.User) ([]byte, error)

	ListCampaignByUser(ctx context.Context, usr uuid.UUID) ([]Campaign, error)
	ListCampaignByOrganisation(ctx context.Context, usr uuid.UUID, org uuid.UUID) ([]Campaign, error)
}

type Auth struct {
	User         uuid.UUID
	Organisation uuid.UUID
}

type CreateCampaignParams struct {
	Schema  string
	Payload []byte
}

type CreateCampaignV1 struct {
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	Organisation string       `json:"organisation,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	Options      *Options     `json:"options"`
	Registration Registration `json:"registration"`
}

type Campaign struct {
	ID           uuid.UUID    `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	Status       string       `json:"status"`
	Organisation string       `json:"organisation,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	Options      *Options     `json:"options"`
	Finance      Finance      `json:"finance,omitempty"`
	Registration Registration `json:"registration"`
}

type Finance struct {
	VoteCurrency string `json:"currency"`
	PlatformFees uint8  `json:"platform_fee_percentage"`
}

type Options struct {
	VoteMetricsVisibility string    `json:"vote_metrics_visibility"`
	ActiviationMethod     string    `json:"activiation_method"`
	VoteStarts            time.Time `json:"vote_starts"`
	VoteEnds              time.Time `json:"vote_ends"`
	VotePerAmount         uint32    `json:"vote_per_amount"`
	AllowRefunds          bool      `json:"allow_refunds"`
}

type Registration struct {
	Gateway      string  `json:"gateway"` // free, paid
	Amount       float32 `json:"amount,omitempty"`
	Requirements []byte  `json:"requirements,omitempty"`
}

type ResolveContestantParams struct {
	RequestID      int64
	Campaign       uuid.UUID
	ContestantUser uuid.UUID
	Verdict        string
}

type CreateRequestContestantParam struct {
	RequestContestantUser gmodels.User
	Organisation          uuid.UUID
	Campaign              uuid.UUID
}

type RequestContestant struct {
	ID             int64
	Campaign       string
	OrganisationID uuid.UUID
	Verdict        string
	Initiated      string
	Time           time.Time
}

type VotePeriod struct {
	Start time.Time
	End   time.Time
}
