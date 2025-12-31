package organisations

import "github.com/google/uuid"

type CreateOrganisationParams struct {
	UserID     uuid.UUID
	Namee      string
	Decription string
}

type AddMember struct {
	ValueType uint8
	Value     string
}
