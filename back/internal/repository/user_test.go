package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func newUserRepoTestSetup(t *testing.T) (*UserRepo, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("create pgx mock: %v", err)
	}
	t.Cleanup(mock.Close)

	repo, err := NewUserRepo(mock)
	if err != nil {
		t.Fatalf("create user repository: %v", err)
	}
	return repo, mock
}

func checkUserRepoExpectations(t *testing.T, mock pgxmock.PgxPoolIface) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet database expectations: %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	userRepo, mock := newUserRepoTestSetup(t)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	input := domain.User{
		TelegramID: 785,
		FirstName:  "Alex",
		LastName:   "Smith",
		Username:   "alexsmith",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	want := input
	want.ID = 42

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(input.TelegramID, input.FirstName, input.LastName, input.Username,
			input.Phone, input.CreatedAt, input.UpdatedAt).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "telegram_id", "first_name", "last_name", "username", "phone", "created_at", "updated_at",
		}).AddRow(
			want.ID, want.TelegramID, want.FirstName, want.LastName, want.Username,
			want.Phone, want.CreatedAt, want.UpdatedAt,
		))

	got, err := userRepo.CreateUser(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if got == nil || *got != want {
		t.Errorf("CreateUser() = %#v, want %#v", got, want)
	}
	checkUserRepoExpectations(t, mock)
}

func TestFindUserByTelegramID(t *testing.T) {
	userRepo, mock := newUserRepoTestSetup(t)
	now := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	want := domain.User{
		ID:         42,
		TelegramID: 785,
		FirstName:  "Alex",
		LastName:   "Smith",
		Username:   "alexsmith",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	mock.ExpectQuery("FROM users WHERE telegram_id").
		WithArgs(want.TelegramID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "telegram_id", "first_name", "last_name", "username", "phone", "created_at", "updated_at",
		}).AddRow(
			want.ID, want.TelegramID, want.FirstName, want.LastName, want.Username,
			want.Phone, want.CreatedAt, want.UpdatedAt,
		))

	got, err := userRepo.FindUserByTelegramID(context.Background(), want.TelegramID)
	if err != nil {
		t.Fatalf("FindUserByTelegramID() error = %v", err)
	}
	if got == nil || *got != want {
		t.Errorf("FindUserByTelegramID() = %#v, want %#v", got, want)
	}
	checkUserRepoExpectations(t, mock)
}

func TestFindUserByTelegramIDNotFoundReturnsNil(t *testing.T) {
	userRepo, mock := newUserRepoTestSetup(t)
	mock.ExpectQuery("FROM users WHERE telegram_id").
		WithArgs(int64(785)).
		WillReturnError(pgx.ErrNoRows)

	got, err := userRepo.FindUserByTelegramID(context.Background(), 785)
	if err != nil {
		t.Fatalf("FindUserByTelegramID() error = %v, want nil", err)
	}
	if got != nil {
		t.Errorf("FindUserByTelegramID() = %#v, want nil", got)
	}
	checkUserRepoExpectations(t, mock)
}
