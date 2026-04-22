package domain

import "errors"

var (
	ErrNotFound         = errors.New("notification not found")
	ErrAlreadySent      = errors.New("notification already sent")
	ErrAlreadyCancelled = errors.New("notification already cancelled")
	ErrInvalidChannel   = errors.New("invalid channel")
	ErrEmptyText        = errors.New("text must not be empty")
	ErrPastPlannedAt    = errors.New("planned_at must be in the future")
	ErrUserNotFound     = errors.New("user not found")
	ErrNoTelegramID     = errors.New("user has no telegram id")
	ErrNoEmail          = errors.New("user has no email")
	ErrNoContact        = errors.New("at least one contact required: telegram_id or email")
)
