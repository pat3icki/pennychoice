package services

import (
	"github.com/pat3icki/pennychoice/internal/services/campaign"
	"github.com/pat3icki/pennychoice/internal/services/organisation"
	"github.com/pat3icki/pennychoice/internal/services/users"
)

type Services struct {
	User         *users.Service
	Campaign     *campaign.Service
	Organisation *organisation.Service
}

type ServicesImpl struct {
}
