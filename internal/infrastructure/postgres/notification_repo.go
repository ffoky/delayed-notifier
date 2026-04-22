package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"

	"DelayedNotifier/internal/domain"
)

var (
	//go:embed sql/queries/notification/save.sql
	saveNotificationQuery string

	//go:embed sql/queries/notification/find_by_id.sql
	findNotificationByIDQuery string

	//go:embed sql/queries/notification/update_status.sql
	updateNotificationStatusQuery string

	//go:embed sql/queries/notification/update_status_with_sent_at.sql
	updateNotificationStatusWithSentAtQuery string

	//go:embed sql/queries/notification/increment_retries.sql
	incrementNotificationRetriesQuery string

	//go:embed sql/queries/notification/find_ready_to_send.sql
	findReadyToSendQuery string

	//go:embed sql/queries/notification/list.sql
	listNotificationsQuery string
)

type NotificationRepo struct {
	db *pgxdriver.Postgres
}

// TODO: добавить контекст
func NewNotificationRepo(db *pgxdriver.Postgres) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Save(ctx context.Context, n *domain.Notification) error {
	_, err := r.db.Exec(ctx, saveNotificationQuery,
		n.ID, n.UserID, string(n.Channel), n.Text, string(n.Status), n.PlannedAt, n.CreatedAt, n.SentAt, n.Retries,
	)
	if err != nil {
		return fmt.Errorf("save notification: %w", err)
	}
	return nil
}

func (r *NotificationRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	row := r.db.QueryRow(ctx, findNotificationByIDQuery, id)
	n, err := scanNotification(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find notification by id: %w", err)
	}
	return n, nil
}

func (r *NotificationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) error {
	if _, err := r.db.Exec(ctx, updateNotificationStatusQuery, string(status), id); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

func (r *NotificationRepo) UpdateStatusWithSentAt(ctx context.Context, id uuid.UUID, status domain.Status, sentAt time.Time) error {
	if _, err := r.db.Exec(ctx, updateNotificationStatusWithSentAtQuery, string(status), sentAt, id); err != nil {
		return fmt.Errorf("update status with sent_at: %w", err)
	}
	return nil
}

// TODO: IncrementRetries это разве не бизнес логика????
func (r *NotificationRepo) IncrementRetries(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.Exec(ctx, incrementNotificationRetriesQuery, id); err != nil {
		return fmt.Errorf("increment retries: %w", err)
	}
	return nil
}

// TODO: это разве не бизнес логиика???
func (r *NotificationRepo) FindReadyToSend(ctx context.Context) ([]*domain.Notification, error) {
	rows, err := r.db.Query(ctx, findReadyToSendQuery)
	if err != nil {
		return nil, fmt.Errorf("find ready to send: %w", err)
	}
	defer rows.Close()

	var (
		notifications []*domain.Notification
		n             *domain.Notification
	)
	for rows.Next() {
		n, err = scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("find ready to send: scan: %w", err)
		}
		notifications = append(notifications, n)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("find ready to send: rows: %w", err)
	}
	return notifications, nil
}

func (r *NotificationRepo) List(ctx context.Context, limit int) ([]*domain.Notification, error) {
	rows, err := r.db.Query(ctx, listNotificationsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var (
		notifications []*domain.Notification
		n             *domain.Notification
	)
	for rows.Next() {
		n, err = scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("list notifications: scan: %w", err)
		}
		notifications = append(notifications, n)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list notifications: rows: %w", err)
	}
	return notifications, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNotification(s scanner) (*domain.Notification, error) {
	n := &domain.Notification{}
	var channel, status string
	var sentAt *time.Time
	if err := s.Scan(&n.ID, &n.UserID, &channel, &n.Text, &status, &n.PlannedAt, &n.CreatedAt, &sentAt, &n.Retries); err != nil {
		return nil, fmt.Errorf("scan notification: %w", err)
	}
	n.Channel = domain.Channel(channel)
	n.Status = domain.Status(status)
	n.SentAt = sentAt
	return n, nil
}
