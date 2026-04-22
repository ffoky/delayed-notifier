package main

import (
	"DelayedNotifier/internal/application"
	"DelayedNotifier/internal/domain"
	"DelayedNotifier/internal/infrastructure/config"
	"DelayedNotifier/internal/infrastructure/postgres"
	"DelayedNotifier/internal/infrastructure/rabbitmq"
	redisinfra "DelayedNotifier/internal/infrastructure/redis"
	"DelayedNotifier/internal/infrastructure/scheduler"
	"DelayedNotifier/internal/infrastructure/sender"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
	"github.com/wb-go/wbf/logger"
	wbfrabbitmq "github.com/wb-go/wbf/rabbitmq"
	wbfredis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	// TODO: использовать zlog.Init() для прода (JSON), zlog.InitConsole() только для локальной разработки
	// переключение через APP_ENV: if os.Getenv("APP_ENV") == "production" { zlog.Init() } else { zlog.InitConsole() }
	zlog.InitConsole()
	if err := run(); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("run failed")
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	appLogger := logger.NewZerologAdapter("delayednotifier-notifier", os.Getenv("APP_ENV"))
	db, err := pgxdriver.New(cfg.Database.URL, appLogger)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()
	rmq := cfg.RabbitMQ

	reconStrat := retry.Strategy{
		Attempts: rmq.ReconnectAttempts,
		Delay:    rmq.ReconnectDelay,
		Backoff:  rmq.Backoff,
	}
	producingStrat := retry.Strategy{
		Attempts: rmq.ProduceAttempts,
		Delay:    rmq.ProduceDelay,
		Backoff:  rmq.Backoff,
	}
	consumingStrat := retry.Strategy{
		Attempts: rmq.ConsumeAttempts,
		Delay:    rmq.ConsumeDelay,
		Backoff:  rmq.Backoff,
	}

	rabbitClient, err := wbfrabbitmq.NewClient(wbfrabbitmq.ClientConfig{
		URL:            rmq.URL,
		ConnectionName: "notifier",
		ReconnectStrat: reconStrat,
		ProducingStrat: producingStrat,
		ConsumingStrat: consumingStrat,
	})
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer func() {
		_ = rabbitClient.Close()
	}()
	if err = rabbitmq.DeclareTopology(rabbitClient); err != nil {
		return fmt.Errorf("declare topology: %w", err)
	}
	redisCache, err := wbfredis.Connect(wbfredis.Options{
		Address:   cfg.Redis.Addr,
		Password:  cfg.Redis.Password,
		MaxMemory: "100mb",
		Policy:    "allkeys-lru",
	})
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("redis connect failed, cache disabled")
	}
	notifRepo := postgres.NewNotificationRepo(db)
	userRepo := postgres.NewUserRepo(db)
	cache := redisinfra.NewNotificationCache(redisCache)
	publisher := rabbitmq.NewNotificationPublisher(rabbitClient)

	senders := []domain.Sender{
		sender.NewTelegramSender(cfg.Telegram.BotToken),
		sender.NewEmailSender(cfg.Email.SMTPHost, cfg.Email.SMTPPort,
			cfg.Email.From, cfg.Email.Username, cfg.Email.Password),
	}

	svc := application.NewService(
		notifRepo,
		userRepo,
		cache,
		publisher,
		senders,
		cfg.App.MaxRetries,
	)

	sched := scheduler.New(svc, cfg.App)
	if err = sched.Start(); err != nil {
		return fmt.Errorf("start scheduler: %w", err)
	}
	defer sched.Stop()

	consumer := rabbitmq.NewNotificationConsumer(rabbitClient, svc)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	zlog.Logger.Info().Msg("notifier starting")
	waitForShutdown(ctx, consumer)
	return nil
}

func waitForShutdown(ctx context.Context, consumer *wbfrabbitmq.Consumer) {
	errCh := make(chan error, 1)
	go func() {
		errCh <- consumer.Start(ctx)
	}()

	select {
	case <-ctx.Done():
		zlog.Logger.Info().Msg("notifier shutting down")
	case err := <-errCh:
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("consumer stopped")
		}
	}
}
