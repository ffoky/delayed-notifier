package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, telegramID *int64, email *string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetOrCreate(ctx context.Context, telegramID *int64, email *string) (*User, error)
}
