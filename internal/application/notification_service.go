package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/domain"
)

const listLimit = 100

type Service struct {
	repo       domain.NotificationRepository
	users      domain.UserRepository
	cache      domain.NotificationCache
	publisher  domain.Publisher
	senders    map[domain.Channel]domain.Sender
	maxRetries int
}

func NewService(
	repo domain.NotificationRepository,
	users domain.UserRepository,
	cache domain.NotificationCache,
	publisher domain.Publisher,
	senders []domain.Sender,
	maxRetries int,
) *Service {
	senderMap := make(map[domain.Channel]domain.Sender, len(senders))
	for _, s := range senders {
		channel := s.Channel()
		senderMap[channel] = s
	}
	return &Service{
		repo:       repo,
		users:      users,
		cache:      cache,
		publisher:  publisher,
		senders:    senderMap,
		maxRetries: maxRetries,
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, channel domain.Channel, text string, plannedAt time.Time) (*domain.Notification, error) {
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	n, err := domain.NewNotification(userID, channel, text, plannedAt)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	if err = s.repo.Save(ctx, n); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	if err = s.cache.Set(ctx, n); err != nil { //TODO: почему тут warn? почему не возвращаем ошибку
		zlog.Logger.Warn().Str("id", n.ID.String()).Err(err).Msg("cache set")
	}

	if err = s.publisher.Publish(ctx, n.ID); err != nil { //TODO: почему тут warn?
		zlog.Logger.Warn().Str("id", n.ID.String()).Err(err).Msg("publish")
	}

	return n, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	if n, err := s.cache.Get(ctx, id); err == nil {
		return n, nil
	}

	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	if err = s.cache.Set(ctx, n); err != nil {
		zlog.Logger.Warn().Str("id", id.String()).Err(err).Msg("cache set")
	}

	return n, nil
}

func (s *Service) Cancel(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cancel: %w", err)
	}

	if err = n.Cancel(); err != nil {
		return nil, fmt.Errorf("cancel: %w", err)
	}

	if err = s.repo.UpdateStatus(ctx, n.ID, n.Status); err != nil {
		return nil, fmt.Errorf("cancel: %w", err)
	}

	if err = s.cache.Delete(ctx, id); err != nil {
		zlog.Logger.Warn().Str("id", id.String()).Err(err).Msg("cache delete")
	}

	return n, nil
}

func (s *Service) Send(ctx context.Context, n *domain.Notification) error {
	if !n.IsReady() {
		return nil
	}

	u, err := s.users.FindByID(ctx, n.UserID)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	recipient, err := resolveRecipient(n.Channel, u)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	sender, ok := s.senders[n.Channel]
	if !ok { //TODO: эту ошибку мб вынести надо в доменные?
		return fmt.Errorf("send: unsupported channel %s", n.Channel)
	}

	if err = sender.Send(ctx, recipient, n.Text); err != nil {
		zlog.Logger.Error().Str("id", n.ID.String()).Err(err).Msg("sender send failed")
		return s.handleSendFailure(ctx, n)
	}

	n.MarkSent()
	if err = s.repo.UpdateStatusWithSentAt(ctx, n.ID, n.Status, *n.SentAt); err != nil {
		return fmt.Errorf("send: update sent: %w", err)
	}

	if err = s.cache.Delete(ctx, n.ID); err != nil { //TODO: почему warn?
		zlog.Logger.Warn().Str("id", n.ID.String()).Err(err).Msg("cache delete after send")
	}

	return nil
}

func (s *Service) List(ctx context.Context) ([]*domain.Notification, error) {
	ns, err := s.repo.List(ctx, listLimit)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	return ns, nil
}

func (s *Service) ProcessReady(ctx context.Context) error {
	notifications, err := s.repo.FindReadyToSend(ctx)
	if err != nil {
		return fmt.Errorf("process ready: %w", err)
	}

	for _, n := range notifications {
		if err = s.publisher.Publish(ctx, n.ID); err != nil {
			zlog.Logger.Error().Str("id", n.ID.String()).Err(err).Msg("publish ready notification")
		}
	}

	return nil
}

func (s *Service) handleSendFailure(ctx context.Context, n *domain.Notification) error {
	if repoErr := s.repo.IncrementRetries(ctx, n.ID); repoErr != nil {
		return fmt.Errorf("send: increment retries: %w", repoErr)
	}
	n.Retries++

	if n.Retries >= s.maxRetries {
		if repoErr := s.repo.UpdateStatus(ctx, n.ID, domain.StatusFailed); repoErr != nil {
			return fmt.Errorf("send: update failed status: %w", repoErr)
		}
		return nil
	}

	if repoErr := s.repo.UpdateStatus(ctx, n.ID, domain.StatusPlanned); repoErr != nil {
		return fmt.Errorf("send: update planned status: %w", repoErr)
	}

	if pubErr := s.publisher.PublishRetry(ctx, n.ID, n.Retries); pubErr != nil {
		zlog.Logger.Warn().Str("id", n.ID.String()).Err(pubErr).Msg("publish retry")
	}

	return nil
}

func resolveRecipient(channel domain.Channel, u *domain.User) (string, error) {
	switch channel {
	case domain.ChannelTelegram:
		if u.TelegramID == nil {
			return "", domain.ErrNoTelegramID
		}
		return strconv.FormatInt(*u.TelegramID, 10), nil
	case domain.ChannelEmail:
		if u.Email == nil {
			return "", domain.ErrNoEmail
		}
		return *u.Email, nil
	default:
		return "", domain.ErrInvalidChannel
	}
}
