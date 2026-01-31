package organisations

import "github.com/google/uuid"

type CreateOrganisationParams struct {
	UserID     uuid.UUID
	Namee      string
	Decription string
	Tags       []string
}

type AddMember struct {
	ValueType uint8
	Value     string
}

type RemoveMember struct {
}

type DeletedOrganisation struct{}
