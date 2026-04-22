package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Channel string

const (
	ChannelTelegram Channel = "telegram"
	ChannelEmail    Channel = "email"
)

type Status string

const (
	StatusPlanned   Status = "planned"
	StatusSending   Status = "sending"
	StatusSent      Status = "sent"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Channel   Channel
	Text      string
	Status    Status
	PlannedAt time.Time
	CreatedAt time.Time
	SentAt    *time.Time
	Retries   int
}

func NewNotification(userID uuid.UUID, channel Channel, text string, plannedAt time.Time) (*Notification, error) {
	if err := validateChannel(channel); err != nil {
		return nil, fmt.Errorf("new notification: %w", err)
	}
	if err := validateText(text); err != nil {
		return nil, fmt.Errorf("new notification: %w", err)
	}
	if err := validatePlannedAt(plannedAt); err != nil {
		return nil, fmt.Errorf("new notification: %w", err)
	}

	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Channel:   channel,
		Text:      text,
		Status:    StatusPlanned,
		PlannedAt: plannedAt.UTC(),
		CreatedAt: time.Now().UTC(),
		Retries:   0,
	}, nil
}

func (n *Notification) Cancel() error {
	switch n.Status {
	case StatusSent:
		return fmt.Errorf("cancel notification: %w", ErrAlreadySent)
	case StatusCancelled:
		return fmt.Errorf("cancel notification: %w", ErrAlreadyCancelled)
	case StatusPlanned, StatusSending, StatusFailed:
		n.Status = StatusCancelled
		return nil
	}
	return nil
}

func (n *Notification) IsReady() bool {
	return n.Status == StatusPlanned && !time.Now().UTC().Before(n.PlannedAt)
}

func (n *Notification) MarkSent() {
	now := time.Now().UTC()
	n.Status = StatusSent
	n.SentAt = &now
}

func validateChannel(ch Channel) error {
	switch ch {
	case ChannelTelegram, ChannelEmail:
		return nil
	default:
		return ErrInvalidChannel
	}
}

func validateText(text string) error {
	if text == "" {
		return ErrEmptyText
	}
	return nil
}

func validatePlannedAt(t time.Time) error {
	if t.Before(time.Now().UTC()) {
		return ErrPastPlannedAt
	}
	return nil
}
