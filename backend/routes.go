package backend

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routes registers all route groups and their handlers.
func (b *Backend) routes(r chi.Router) {
	// Global middleware (logging, recovery, etc.)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes (no auth)
	r.Group(func(r chi.Router) {
		// Health check, etc.
		r.Get("/health", b.handleHealth)
		r.Get("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "backend/swagger.yaml")
		})
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(b.UserAuthorize) // injects client ID into context

		// Users
		r.Route("/users", func(r chi.Router) {
			r.Post("/", b.handleCreateUser)
			r.Post("/contestant", b.handleCreateContestantProfile)
			r.Post("/login", b.handleLoginUser)
			// r.Get("/profile", b.handleGetUserProfile)
			// r.Post("/deactivate/init", b.handleInitDeactivation)
			// r.Post("/deactivate/confirm", b.handleDeactivateUser)
			// Password reset endpoints, etc.
		})

		// Campaigns
		r.Route("/campaigns", func(r chi.Router) {
			r.Post("/", b.handleCreateCampaign)
			r.Post("/activate", b.handleActivateCampaign)
			r.Post("/{id}/activate", b.handleActivateCampaign)
			r.Post("/pause", b.handlePauseCampaign)
			r.Post("/{id}/pause", b.handlePauseCampaign)
			r.Post("/vote-period", b.handleUpdateVotePeriodCampaign)
			r.Post("/{id}/vote-period", b.handleUpdateVotePeriodCampaign)
		})

		// Organisations
		r.Route("/organisations", func(r chi.Router) {
			// ...
		})
	})
}

// handleHealth is a simple health check endpoint.
func (b *Backend) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, uuid.Nil, map[string]string{"status": "ok"}, http.StatusOK)
}
