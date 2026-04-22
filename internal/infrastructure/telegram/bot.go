package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/domain"
)

type Bot struct {
	api *tgbotapi.BotAPI
	svc domain.UserService
}

func NewBot(token string, svc domain.UserService) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("new bot api: %w", err)
	}
	return &Bot{api: api, svc: svc}, nil
}

func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			if update.Message.Command() == "start" {
				b.handleStart(ctx, update.Message)
			}
		}
	}
}

func (b *Bot) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	u, err := b.svc.Create(ctx, &chatID, nil)
	if err != nil {
		zlog.Logger.Error().Err(err).Int64("chat_id", chatID).Msg("create user from telegram")
		b.reply(msg, "Произошла ошибка. Попробуйте позже.")
		return
	}

	zlog.Logger.Info().Str("user_id", u.ID.String()).Int64("chat_id", chatID).Msg("user created via telegram")
	b.reply(msg, fmt.Sprintf("Готово! Ваш ID: %s\nИспользуйте его для создания уведомлений.", u.ID.String()))
}

func (b *Bot) reply(msg *tgbotapi.Message, text string) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	if _, err := b.api.Send(reply); err != nil {
		zlog.Logger.Error().Err(err).Msg("send telegram reply")
	}
}
