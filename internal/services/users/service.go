package users

import (
	"github.com/pat3icki/pennychoice/internal/db/redis"
)

type Service struct {
	PostgresSQL SQL
	Redis       redis.Client
}
