package organisations

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/pat3icki/pennychoice/internal/db/sqlc"
	"github.com/pat3icki/pennychoice/types"
)

type Service struct {
	PostgreSQL sqlc.Queries
}

func (s Service) CreateOrganisation(ctx context.Context, creat_usr types.User, create_org *CreateOrganisationParams) (uuid.UUID, error) {
	var id uuid.UUID
	if creat_usr.Uniqueness != types.Unique_ID {
		return uuid.Nil, errors.ErrUnsupported
	}
	s.PostgreSQL.CreateOrganisation(ctx, sqlc.CreateOrganisationParams{
		ID:            uuid.New(),
		Name:          create_org.Name,
		CreatorUserID: uuid.MustParse(creat_usr.Value),
	})
	return id, nil

}

func (s Service) UpdateUser(initUsr types.User, usrs types.User, action uint8) error

func (s Service) UpdatePermissionData(initUsr types.User, data json.RawMessage) error

func (s Service) GetPermissionData()

func (s *Service) CreateDecison()
func (s *Service) Excuteecison()
func (s *Service) GetDecisonInfo()
