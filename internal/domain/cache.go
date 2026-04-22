package domain

//go:generate mockgen -source=cache.go -destination=mocks/mock_notification_cache.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
)

type NotificationCache interface {
	Get(ctx context.Context, id uuid.UUID) (*Notification, error)
	Set(ctx context.Context, n *Notification) error
	Delete(ctx context.Context, id uuid.UUID) error
}
