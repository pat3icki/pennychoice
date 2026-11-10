package messenger

import (
	"context"
	"sync"
	"time"
)

type Medium string

const (
	Email    Medium = "email"
	Whatsapp Medium = "whatsapp_phone"
	SMS      Medium = "sms"
)

type Client struct {
	mu  sync.Mutex
	OTP OTP
}

type OTP interface {
	Create(ctx context.Context, id string, mode Medium, opts []byte, expire time.Time) error
	Validate(ctx context.Context, id string, code string) error
}
