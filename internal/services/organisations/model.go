package organisations

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateOrganisationParams struct {
	Name                 string
	Decription           string
	Tags                 []string
	MaximumOrganisers    uint8
	MaximumActiveCampign uint8
	Permissions          json.RawMessage
}

type Organisation struct {
	ID          uuid.UUID
	Name        string
	Description string
	Tags        []string
	Permissions json.RawMessage
	CreatedAt   time.Time
}

type DecisonInfo struct {
	ID             uuid.UUID
	CampaignID     uuid.UUID
	OrganisationID uuid.UUID
	ActionType     string
	Message        string
	Deadline       time.Time
	Status         string
}

type Config struct {
	MaximumOrganisers    uint
	MaximumActiveCampign uint
}

type CreateDecisonParams struct {
	ActionType     string
	CampaignID     uuid.UUID
	OrganisationID uuid.UUID
	Message        string
	Init           uuid.UUID
	Deadline       time.Time
}
