package scheduler

import (
	"context"
	"fmt"

	"github.com/go-co-op/gocron/v2"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/domain"
	"DelayedNotifier/internal/infrastructure/config"
)

type Scheduler struct {
	svc    domain.NotificationService
	cfg    config.AppConfig
	gocron gocron.Scheduler
}

func New(svc domain.NotificationService, cfg config.AppConfig) *Scheduler {
	return &Scheduler{
		svc: svc,
		cfg: cfg,
	}
}

func (s *Scheduler) Start() error {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("create new cron scheduler: %w", err)
	}

	_, err = scheduler.NewJob(
		gocron.DurationJob(s.cfg.SchedulerInterval),
		gocron.NewTask(s.processReady),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		return fmt.Errorf("schedule process ready job: %w", err)
	}

	s.gocron = scheduler
	scheduler.Start()
	return nil
}

func (s *Scheduler) Stop() {
	if s.gocron != nil {
		if err := s.gocron.Shutdown(); err != nil {
			zlog.Logger.Error().Err(err).Msg("scheduler shutdown")
		}
	}
}

func (s *Scheduler) processReady() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ProcessReadyTimeout)
	defer cancel()

	if err := s.svc.ProcessReady(ctx); err != nil {
		zlog.Logger.Error().Err(err).Msg("process ready notifications")
	}
}
