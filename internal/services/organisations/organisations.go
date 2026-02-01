package organisations

import (
	"context"

	"github.com/pat3icki/pennychoice/types"
)

type Service struct {
}

func (s Service) CreateOrganisation(ctx context.Context, create_usr types.User, create_org *CreateOrganisationParams)

func (s Service) MakeUserAdmin(ctx context.Context, adm_usr types.User)

func (s Service) RemoveUserAdmin()

func (s Service) AddMember()

func (s Service) RemoveMember()
