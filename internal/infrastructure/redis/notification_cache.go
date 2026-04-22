package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	wbfredis "github.com/wb-go/wbf/redis"

	"DelayedNotifier/internal/domain"
)

const cacheTTL = 5 * time.Minute

type NotificationCache struct {
	client *wbfredis.Client
}

func NewNotificationCache(client *wbfredis.Client) *NotificationCache {
	return &NotificationCache{client: client}
}

func (c *NotificationCache) Get(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	val, err := c.client.Get(ctx, cacheKey(id))
	if err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	}

	var dto notificationDTO
	if err = json.Unmarshal([]byte(val), &dto); err != nil {
		return nil, fmt.Errorf("cache get: unmarshal: %w", err)
	}

	n, err := fromDTO(dto)
	if err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	}

	return n, nil
}

func (c *NotificationCache) Set(ctx context.Context, n *domain.Notification) error {
	data, err := json.Marshal(toDTO(n))
	if err != nil {
		return fmt.Errorf("cache set: marshal: %w", err)
	}

	if err = c.client.SetWithExpiration(ctx, cacheKey(n.ID), string(data), cacheTTL); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	return nil
}

func (c *NotificationCache) Delete(ctx context.Context, id uuid.UUID) error {
	if err := c.client.Del(ctx, cacheKey(id)); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}

	return nil
}

func cacheKey(id uuid.UUID) string {
	return "notification:" + id.String()
}
