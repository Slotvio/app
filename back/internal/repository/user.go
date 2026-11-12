package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	db DB
}

func NewUserRepo(db DB) (*UserRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("user repository: db is nil")
	}
	return &UserRepo{db: db}, nil
}

const userColumns = `
	id,
	telegram_id,
	first_name,
	COALESCE(last_name, ''),
	COALESCE(username, ''),
	COALESCE(phone, ''),
	created_at,
	updated_at`


func (r *UserRepo) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (
			telegram_id,
			first_name,
			last_name,
			username,
			phone,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			NULLIF($4, ''),
			NULLIF($5, ''),
			$6,
			$7
		)
		ON CONFLICT (telegram_id) DO UPDATE
		SET first_name = EXCLUDED.first_name,
		    last_name = EXCLUDED.last_name,
		    username = EXCLUDED.username,
		    updated_at = EXCLUDED.updated_at
		RETURNING ` + userColumns

	created, err := scanUser(r.db.QueryRow(ctx, query,
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Phone,
		user.CreatedAt,
		user.UpdatedAt,
	))
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &created, nil
}

func (r *UserRepo) FindUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE telegram_id = $1`
	user, err := scanUser(r.db.QueryRow(ctx, query, telegramID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by telegram id: %w", err)
	}
	return &user, nil
}

func scanUser(row pgx.Row) (domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}
