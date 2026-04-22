package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type NotificationService interface {
	Create(ctx context.Context, userID uuid.UUID, channel Channel, text string, plannedAt time.Time) (*Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	Cancel(ctx context.Context, id uuid.UUID) (*Notification, error)
	Send(ctx context.Context, n *Notification) error
	List(ctx context.Context) ([]*Notification, error)
	ProcessReady(ctx context.Context) error
}
