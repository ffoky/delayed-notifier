package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/google/uuid"
	wbfrabbitmq "github.com/wb-go/wbf/rabbitmq"
)

const (
	mainExchange = "notify.exchange"
	mainQueue    = "notifications"
	contentType  = "text/plain"

	retryQueue5s  = "notifications.retry.5s"
	retryQueue30s = "notifications.retry.30s"
	retryQueue2m  = "notifications.retry.2m"

	secondRetryCount = 2
)

type NotificationPublisher struct {
	main  *wbfrabbitmq.Publisher
	retry *wbfrabbitmq.Publisher
}

func NewNotificationPublisher(client *wbfrabbitmq.RabbitClient) *NotificationPublisher {
	return &NotificationPublisher{
		main:  wbfrabbitmq.NewPublisher(client, mainExchange, contentType),
		retry: wbfrabbitmq.NewPublisher(client, amqp.DefaultExchange, contentType),
	}
}

func (p *NotificationPublisher) Publish(ctx context.Context, notificationID uuid.UUID) error {
	if err := p.main.Publish(ctx, []byte(notificationID.String()), mainQueue); err != nil {
		return fmt.Errorf("publish notification: %w", err)
	}
	return nil
}

func (p *NotificationPublisher) PublishRetry(ctx context.Context, notificationID uuid.UUID, retryCount int) error {
	queue := retryQueueName(retryCount)
	if err := p.retry.Publish(ctx, []byte(notificationID.String()), queue); err != nil {
		return fmt.Errorf("publish retry notification: %w", err)
	}
	return nil
}

func retryQueueName(retryCount int) string {
	switch retryCount {
	case 1:
		return retryQueue5s
	case secondRetryCount:
		return retryQueue30s
	default:
		return retryQueue2m
	}
}
