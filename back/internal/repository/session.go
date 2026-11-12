package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/jackc/pgx/v5"
)

type SessionRepo struct {
	db DB
}

func NewSessionRepo(db DB) (*SessionRepo, error) {
	if db == nil {
		return nil, fmt.Errorf("session repository: db is nil")
	}
	return &SessionRepo{db: db}, nil
}

func (r *SessionRepo) CreateSession(ctx context.Context, session domain.AuthSession) (*domain.AuthSession, error) {
	const query = `
		INSERT INTO auth_sessions (
			user_id,
			token_hash,
			created_at,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, token_hash, created_at, expires_at;
	`

	created, err := scanSession(r.db.QueryRow(ctx, query,
		session.UserID,
		session.TokenHash[:],
		session.CreatedAt,
		session.ExpiresAt,
	))
	if err != nil {
		return nil, fmt.Errorf("create auth session: %w", err)
	}
	return &created, nil
}

func (r *SessionRepo) FindSessionByUserID(ctx context.Context, userID int64) (*domain.AuthSession, error) {
	const query = `
		SELECT id, user_id, token_hash, created_at, expires_at
		FROM auth_sessions
		WHERE user_id = $1
		  AND expires_at > NOW()
		  AND revoked_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT 1;
	`

	session, err := scanSession(r.db.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find active session by user id: %w", err)
	}
	return &session, nil
}

func (r *SessionRepo) FindSessionByToken(ctx context.Context, hashToken [32]byte) (int64, error) {
	const query = `
		SELECT user_id
		FROM auth_sessions
		WHERE token_hash = $1
		  AND expires_at > NOW()
		  AND revoked_at IS NULL;
	`

	var userID int64
	if err := r.db.QueryRow(ctx, query, hashToken[:]).Scan(&userID); err != nil {
		return 0, fmt.Errorf("find active session by token hash: %w", err)
	}
	return userID, nil
}

func scanSession(row pgx.Row) (domain.AuthSession, error) {
	var session domain.AuthSession
	var tokenHash []byte

	if err := row.Scan(
		&session.ID,
		&session.UserID,
		&tokenHash,
		&session.CreatedAt,
		&session.ExpiresAt,
	); err != nil {
		return domain.AuthSession{}, err
	}
	if len(tokenHash) != len(session.TokenHash) {
		return domain.AuthSession{}, fmt.Errorf("invalid auth session token hash length: got %d, want %d", len(tokenHash), len(session.TokenHash))
	}
	copy(session.TokenHash[:], tokenHash)
	return session, nil
}
