-- +goose Up
CREATE TABLE users (
                       id          UUID PRIMARY KEY,
                       telegram_id BIGINT,
                       email       TEXT,
                       created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notifications (
                               id         UUID        PRIMARY KEY,
                               user_id    UUID        NOT NULL REFERENCES users(id),
                               channel    TEXT        NOT NULL,
                               text       TEXT        NOT NULL,
                               status     TEXT        NOT NULL,
                               planned_at TIMESTAMPTZ NOT NULL,
                               created_at TIMESTAMPTZ NOT NULL,
                               sent_at    TIMESTAMPTZ,
                               retries    INT         NOT NULL
);

CREATE INDEX idx_notifications_ready ON notifications (planned_at)
    WHERE status = 'planned';

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS users;