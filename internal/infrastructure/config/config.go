package config

import (
	"errors"
	"fmt"
	"time"

	cleanenvport "github.com/wb-go/wbf/config/cleanenv-port"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Database DatabaseConfig `yaml:"database"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Redis    RedisConfig    `yaml:"redis"`
	Telegram TelegramConfig `yaml:"telegram"`
	Email    EmailConfig    `yaml:"email"`
	App      AppConfig      `yaml:"app"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr" env:"HTTP_ADDR" env-default:":8080"`
}

type DatabaseConfig struct {
	URL string `yaml:"url" env:"DATABASE_URL"`
}

type RabbitMQConfig struct {
	URL               string        `yaml:"url"                env:"RABBITMQ_URL"                env-default:"amqp://guest:guest@localhost:5672/"`
	ReconnectAttempts int           `yaml:"reconnect_attempts" env:"RABBITMQ_RECONNECT_ATTEMPTS" env-default:"10"`
	ReconnectDelay    time.Duration `yaml:"reconnect_delay"    env:"RABBITMQ_RECONNECT_DELAY"    env-default:"500ms"`
	ProduceAttempts   int           `yaml:"produce_attempts"   env:"RABBITMQ_PRODUCE_ATTEMPTS"   env-default:"3"`
	ProduceDelay      time.Duration `yaml:"produce_delay"      env:"RABBITMQ_PRODUCE_DELAY"      env-default:"100ms"`
	ConsumeAttempts   int           `yaml:"consume_attempts"   env:"RABBITMQ_CONSUME_ATTEMPTS"   env-default:"5"`
	ConsumeDelay      time.Duration `yaml:"consume_delay"      env:"RABBITMQ_CONSUME_DELAY"      env-default:"1s"`
	Backoff           float64       `yaml:"backoff"            env:"RABBITMQ_BACKOFF"            env-default:"2"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"     env:"REDIS_ADDR"     env-default:"localhost:6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token" env:"TELEGRAM_BOT_API_TOKEN" env-default:""`
}

type EmailConfig struct {
	SMTPHost string `yaml:"smtp_host" env:"EMAIL_SMTP_HOST"`
	SMTPPort int    `yaml:"smtp_port" env:"EMAIL_SMTP_PORT" env-default:"587"`
	From     string `yaml:"from"      env:"EMAIL_FROM"`
	Username string `yaml:"username"  env:"EMAIL_USERNAME"`
	Password string `yaml:"password"  env:"EMAIL_PASSWORD"`
}

type AppConfig struct {
	MaxRetries          int           `yaml:"max_retries"           env:"MAX_RETRIES"           env-default:"5"`
	SchedulerInterval   time.Duration `yaml:"scheduler_interval"    env:"SCHEDULER_INTERVAL"    env-default:"30s"`
	ProcessReadyTimeout time.Duration `yaml:"process_ready_timeout" env:"PROCESS_READY_TIMEOUT" env-default:"25s"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenvport.Load(&cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if cfg.Database.URL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	return &cfg, nil
}
