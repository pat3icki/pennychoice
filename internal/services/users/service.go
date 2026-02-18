package users

import (
	"sync"

	"github.com/pat3icki/pennychoice/internal/db/sqlc"
)

type Service struct {
	mu                sync.RWMutex
	PostgreSQL        *sqlc.Queries
	Redis             Cache
	UserAnyIdentifier bool
}
