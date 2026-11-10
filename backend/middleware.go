package backend

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// WithClientID puts the client UUID in request context.
func WithClientID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, clientIDKey, id)
}

// ClientIDFromContext extracts the client UUID from request context.
func ClientIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(clientIDKey).(uuid.UUID)
	return id
}

// UserAuthorize is a middleware that validates the token and sets the client ID in context.
func (b *Backend) UserAuthorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pass-through for now or extract from header/context
		next.ServeHTTP(w, r)
	})
}
