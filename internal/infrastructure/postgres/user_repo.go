package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	pgxdriver "github.com/wb-go/wbf/dbpg/pgx-driver"

	"DelayedNotifier/internal/domain"
)

var (
	//go:embed sql/queries/user/find_by_id.sql
	findUserByIDQuery string

	//go:embed sql/queries/user/create.sql
	createUserQuery string

	//go:embed sql/queries/user/find_by_email.sql
	findUserByEmailQuery string

	//go:embed sql/queries/user/find_by_telegram_id.sql
	findUserByTelegramIDQuery string
)

type UserRepo struct {
	db *pgxdriver.Postgres
}

func NewUserRepo(db *pgxdriver.Postgres) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.db.QueryRow(ctx, findUserByIDQuery, id)

	u := &domain.User{}
	if err := row.Scan(&u.ID, &u.TelegramID, &u.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, createUserQuery,
		u.ID, u.TelegramID, u.Email,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, findUserByEmailQuery, email)
	u := &domain.User{}
	if err := row.Scan(&u.ID, &u.TelegramID, &u.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) FindByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	row := r.db.QueryRow(ctx, findUserByTelegramIDQuery, telegramID)
	u := &domain.User{}
	if err := row.Scan(&u.ID, &u.TelegramID, &u.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by telegram id: %w", err)
	}
	return u, nil
}
