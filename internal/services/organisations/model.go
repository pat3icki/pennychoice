package organisations

import (
	"encoding/json"
)

type CreateOrganisationParams struct {
	Namee       string
	Decription  string
	Tags        []string
	Permissions json.RawMessage
}

type AddMember struct {
	ValueType uint8
	Value     string
}

type RemoveMember struct {
}

type DeletedOrganisation struct{}
