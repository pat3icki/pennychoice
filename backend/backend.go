package backend

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/pat3icki/pennychoice/internal/config"
	"github.com/pat3icki/pennychoice/internal/db"
	"github.com/pat3icki/pennychoice/internal/services"
	"github.com/pat3icki/pennychoice/internal/services/campaign"
	"github.com/pat3icki/pennychoice/internal/services/users"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const clientIDKey contextKey = "clientID"

// Backend holds the HTTP server, service managers, and configuration.
type Backend struct {
	mu     sync.Mutex
	router chi.Router
	db     *db.Client
	config config.Config
	// UsersServices    *users.Service
	// CampaignServices *campaign.Service
	services services.Services
	// Other managers as needed
}

// New creates a new Backend instance with all dependencies wired.
func New(cfg config.Config, dbClient *db.Client) *Backend {
	sv := services.Services{
		User: &users.Service{
			Postgres: dbClient,
			// Pass any other dependencies (e.g., messenger)
		},
		Campaign: &campaign.Service{
			Postgres: dbClient,
		},
	}
	b := &Backend{
		router:   chi.NewRouter(),
		db:       dbClient,
		config:   cfg,
		services: sv,
	}
	b.routes(b.router)
	return b
}

// Start runs the HTTP server on the given address.
func (b *Backend) Start(addr string) error {
	return http.ListenAndServe(addr, b.router)
}
