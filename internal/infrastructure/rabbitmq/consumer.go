package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	wbfrabbitmq "github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/domain"
)

func NewNotificationConsumer(client *wbfrabbitmq.RabbitClient, svc domain.NotificationService) *wbfrabbitmq.Consumer {
	handler := func(ctx context.Context, d amqp.Delivery) error {
		zlog.Logger.Info().Str("body", string(d.Body)).Msg("consumer received message")
		id, err := uuid.Parse(string(d.Body))
		if err != nil {
			zlog.Logger.Warn().Str("body", string(d.Body)).Err(err).Msg("parse notification id")
			return nil
		}

		n, err := svc.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}
			return fmt.Errorf("consumer get notification: %w", err)
		}

		if n.Status == domain.StatusCancelled || n.Status == domain.StatusSent || n.Status == domain.StatusFailed {
			return nil
		}

		if err = svc.Send(ctx, n); err != nil {
			zlog.Logger.Error().Str("id", id.String()).Err(err).Msg("send failed")
			return fmt.Errorf("consumer send: %w", err)
		}
		return nil
	}

	consumerCfg := wbfrabbitmq.ConsumerConfig{
		Queue:         mainQueue,
		ConsumerTag:   "notifier",
		AutoAck:       false,
		Workers:       1,
		PrefetchCount: 1,
		Nack: wbfrabbitmq.NackConfig{
			Multiple: false,
			Requeue:  false,
		},
	}

	return wbfrabbitmq.NewConsumer(client, consumerCfg, handler)
}
