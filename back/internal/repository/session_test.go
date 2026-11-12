package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func newSessionRepoTestSetup(t *testing.T) (*SessionRepo, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("create pgx mock: %v", err)
	}
	t.Cleanup(mock.Close)

	repo, err := NewSessionRepo(mock)
	if err != nil {
		t.Fatalf("create session repository: %v", err)
	}
	return repo, mock
}

func checkSessionRepoExpectations(t *testing.T, mock pgxmock.PgxPoolIface) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet database expectations: %v", err)
	}
}

func TestCreateSession(t *testing.T) {
	sessionRepo, mock := newSessionRepoTestSetup(t)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	input := domain.AuthSession{
		UserID:    42,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	for i := range input.TokenHash {
		input.TokenHash[i] = byte(i + 1)
	}
	want := input
	want.ID = 10

	mock.ExpectQuery("INSERT INTO auth_sessions").
		WithArgs(input.UserID, input.TokenHash[:], input.CreatedAt, input.ExpiresAt).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "token_hash", "created_at", "expires_at",
		}).AddRow(
			want.ID, want.UserID, want.TokenHash[:], want.CreatedAt, want.ExpiresAt,
		))

	got, err := sessionRepo.CreateSession(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if got == nil || *got != want {
		t.Errorf("CreateSession() = %#v, want %#v", got, want)
	}
	checkSessionRepoExpectations(t, mock)
}

func TestFindSessionByUserID(t *testing.T) {
	sessionRepo, mock := newSessionRepoTestSetup(t)
	want := domain.AuthSession{
		ID:        10,
		UserID:    42,
		CreatedAt: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		ExpiresAt: time.Date(2026, time.January, 3, 3, 4, 5, 0, time.UTC),
	}
	for i := range want.TokenHash {
		want.TokenHash[i] = byte(i + 1)
	}

	mock.ExpectQuery("FROM auth_sessions").
		WithArgs(want.UserID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "token_hash", "created_at", "expires_at",
		}).AddRow(
			want.ID, want.UserID, want.TokenHash[:], want.CreatedAt, want.ExpiresAt,
		))

	got, err := sessionRepo.FindSessionByUserID(context.Background(), want.UserID)
	if err != nil {
		t.Fatalf("FindSessionByUserID() error = %v", err)
	}
	if got == nil || *got != want {
		t.Errorf("FindSessionByUserID() = %#v, want %#v", got, want)
	}
	checkSessionRepoExpectations(t, mock)
}

func TestFindSessionByToken(t *testing.T) {
	sessionRepo, mock := newSessionRepoTestSetup(t)
	var hash [32]byte
	for i := range hash {
		hash[i] = byte(i + 1)
	}
	const wantUserID int64 = 42

	mock.ExpectQuery("WHERE token_hash = \\$1").
		WithArgs(hash[:]).
		WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(wantUserID))

	got, err := sessionRepo.FindSessionByToken(context.Background(), hash)
	if err != nil {
		t.Fatalf("FindSessionByToken() error = %v", err)
	}
	if got != wantUserID {
		t.Errorf("FindSessionByToken() = %d, want %d", got, wantUserID)
	}
	checkSessionRepoExpectations(t, mock)
}

func TestFindSessionByTokenPropagatesNotFound(t *testing.T) {
	sessionRepo, mock := newSessionRepoTestSetup(t)
	var hash [32]byte
	mock.ExpectQuery("WHERE token_hash = \\$1").
		WithArgs(hash[:]).
		WillReturnError(pgx.ErrNoRows)

	_, err := sessionRepo.FindSessionByToken(context.Background(), hash)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("FindSessionByToken() error = %v, want wrapped pgx.ErrNoRows", err)
	}
	checkSessionRepoExpectations(t, mock)
}
