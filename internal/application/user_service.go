package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"DelayedNotifier/internal/domain"
)

type UserService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Create(ctx context.Context, telegramID *int64, email *string) (*domain.User, error) {
	if telegramID == nil && email == nil {
		return nil, fmt.Errorf("create user: %w", domain.ErrNoContact)
	}

	u := &domain.User{
		ID:         uuid.New(),
		TelegramID: telegramID,
		Email:      email,
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return u, nil
}

func (s *UserService) GetOrCreate(ctx context.Context, telegramID *int64, email *string) (*domain.User, error) {
	if telegramID == nil && email == nil {
		return nil, fmt.Errorf("get or create user: %w", domain.ErrNoContact)
	}

	if telegramID != nil {
		u, err := s.users.FindByTelegramID(ctx, *telegramID)
		if err == nil {
			return u, nil
		}
		if !errors.Is(err, domain.ErrUserNotFound) {
			return nil, fmt.Errorf("get or create user: %w", err)
		}
	}

	if email != nil {
		u, err := s.users.FindByEmail(ctx, *email)
		if err == nil {
			return u, nil
		}
		if !errors.Is(err, domain.ErrUserNotFound) {
			return nil, fmt.Errorf("get or create user: %w", err)
		}
	}

	return s.Create(ctx, telegramID, email)
}
