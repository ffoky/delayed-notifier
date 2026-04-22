package redis

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"DelayedNotifier/internal/domain"
)

type notificationDTO struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Channel   string     `json:"channel"`
	Text      string     `json:"text"`
	Status    string     `json:"status"`
	PlannedAt time.Time  `json:"planned_at"`
	CreatedAt time.Time  `json:"created_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	Retries   int        `json:"retries"`
}

func toDTO(n *domain.Notification) notificationDTO {
	return notificationDTO{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		Channel:   string(n.Channel),
		Text:      n.Text,
		Status:    string(n.Status),
		PlannedAt: n.PlannedAt,
		CreatedAt: n.CreatedAt,
		SentAt:    n.SentAt,
		Retries:   n.Retries,
	}
}

func fromDTO(dto notificationDTO) (*domain.Notification, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("parse id: %w", err)
	}

	userID, err := uuid.Parse(dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user_id: %w", err)
	}

	return &domain.Notification{
		ID:        id,
		UserID:    userID,
		Channel:   domain.Channel(dto.Channel),
		Text:      dto.Text,
		Status:    domain.Status(dto.Status),
		PlannedAt: dto.PlannedAt,
		CreatedAt: dto.CreatedAt,
		SentAt:    dto.SentAt,
		Retries:   dto.Retries,
	}, nil
}
