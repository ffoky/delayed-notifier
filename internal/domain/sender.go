package domain

//go:generate mockgen -source=sender.go -destination=mocks/mock_sender.go -package=mocks

import "context"

type Sender interface {
	Send(ctx context.Context, recipient string, text string) error
	Channel() Channel
}
