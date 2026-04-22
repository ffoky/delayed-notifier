package domain

//go:generate mockgen -source=notification_repository.go -destination=mocks/mock_notification_repository.go -package=mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	Save(ctx context.Context, n *Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	UpdateStatusWithSentAt(ctx context.Context, id uuid.UUID, status Status, sentAt time.Time) error
	IncrementRetries(ctx context.Context, id uuid.UUID) error
	FindReadyToSend(ctx context.Context) ([]*Notification, error)
	List(ctx context.Context, limit int) ([]*Notification, error)
}
