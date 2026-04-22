package domain

import (
	"context"

	"github.com/google/uuid"
)

type Publisher interface {
	Publish(ctx context.Context, notificationID uuid.UUID) error
	PublishRetry(ctx context.Context, notificationID uuid.UUID, retryCount int) error
}
