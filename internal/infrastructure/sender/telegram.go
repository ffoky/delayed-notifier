package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/domain"
)

const (
	timeout = 10
)

type TelegramSender struct {
	botToken   string
	httpClient *http.Client
}

func NewTelegramSender(botToken string) *TelegramSender {
	return &TelegramSender{
		botToken:   botToken,
		httpClient: &http.Client{Timeout: timeout * time.Second},
	}
}

func (s *TelegramSender) Send(ctx context.Context, recipient string, text string) error {
	zlog.Logger.Info().Str("recipient", recipient).Msg("sending telegram message")
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)

	payload := map[string]string{
		"chat_id": recipient,
		"text":    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram send: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram send: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send: do request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram send: unexpected status %d", resp.StatusCode)
	}

	return nil
}

func (s *TelegramSender) Channel() domain.Channel {
	return domain.ChannelTelegram
}
