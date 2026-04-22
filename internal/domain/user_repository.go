package domain

//go:generate mockgen -source=user_repository.go -destination=mocks/mock_user_repository.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	Create(ctx context.Context, u *User) error
}
