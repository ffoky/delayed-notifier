package main

import (
	"fmt"
	"os"

	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
	wbfrabbitmq "github.com/wb-go/wbf/rabbitmq"
	wbfredis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/application"
	"DelayedNotifier/internal/domain"
	"DelayedNotifier/internal/infrastructure/config"
	httphandler "DelayedNotifier/internal/infrastructure/http"
	"DelayedNotifier/internal/infrastructure/http/generated"
	"DelayedNotifier/internal/infrastructure/postgres"
	"DelayedNotifier/internal/infrastructure/rabbitmq"
	redisinfra "DelayedNotifier/internal/infrastructure/redis"
	"DelayedNotifier/internal/infrastructure/sender"
)

func main() {
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

	appLogger := logger.NewZerologAdapter("delayednotifier-webserver", os.Getenv("APP_ENV"))

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

	rabbitClient, err := wbfrabbitmq.NewClient(wbfrabbitmq.ClientConfig{
		URL:            rmq.URL,
		ConnectionName: "webserver",
		ReconnectStrat: reconStrat,
		ProducingStrat: producingStrat,
	})
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer func() {
		_ = rabbitClient.Close()
	}()

	rc := wbfredis.New(cfg.Redis.Addr, cfg.Redis.Password, 0)
	notifRepo := postgres.NewNotificationRepo(db)
	userRepo := postgres.NewUserRepo(db)
	cache := redisinfra.NewNotificationCache(rc)
	publisher := rabbitmq.NewNotificationPublisher(rabbitClient)

	senders := []domain.Sender{
		sender.NewTelegramSender(cfg.Telegram.BotToken),
		sender.NewEmailSender(cfg.Email.SMTPHost, cfg.Email.SMTPPort,
			cfg.Email.From, cfg.Email.Username, cfg.Email.Password),
	}

	notificationSvc := application.NewService(notifRepo, userRepo, cache, publisher, senders, cfg.App.MaxRetries)
	userSvc := application.NewUserService(userRepo)
	handler := httphandler.NewHandler(notificationSvc, userSvc)

	engine := ginext.New("release")
	engine.Use(ginext.Logger(), ginext.Recovery())
	generated.RegisterHandlers(engine.Engine, handler)
	engine.Static("/static", "./frontend/static")
	engine.StaticFile("/", "./frontend/index.html")

	zlog.Logger.Info().Str("addr", cfg.HTTP.Addr).Msg("webserver starting")

	if err = engine.Run(cfg.HTTP.Addr); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}
