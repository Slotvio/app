package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/pashagolub/pgxmock/v5"
)

var (
	masterTestTime = time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	masterTest     = domain.Master{
		ID:                9,
		UserID:            12,
		Slug:              "alex",
		Title:             "Alex",
		Category:          "barber",
		Description:       "Description",
		Timezone:          "Europe/Minsk",
		SubscriptionUntil: masterTestTime.AddDate(0, 1, 0),
		CreatedAt:         masterTestTime,
		UpdatedAt:         masterTestTime,
	}
	serviceTest = domain.Service{
		ID:              21,
		MasterID:        9,
		Name:            "Haircut",
		Description:     "Short description",
		DurationMinutes: 60,
		Price:           5000,
		IsActive:        true,
		CreatedAt:       masterTestTime,
		UpdatedAt:       masterTestTime,
	}
	scheduleTest = domain.SchedulePeriod{
		ID:        31,
		MasterID:  9,
		DayOfWeek: 1,
		StartTime: time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC),
		CreatedAt: masterTestTime,
		UpdatedAt: masterTestTime,
	}
)

func newMockRepo(t *testing.T) (*MasterRepo, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("create pgx mock: %v", err)
	}
	t.Cleanup(mock.Close)

	repo, err := NewMasterRepo(mock)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}
	return repo, mock
}

func assertMockExpectations(t *testing.T, mock pgxmock.PgxPoolIface) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet database expectations: %v", err)
	}
}

func masterRows(m domain.Master) *pgxmock.Rows {
	return pgxmock.NewRows([]string{
		"id", "user_id", "slug", "title", "category", "description",
		"timezone", "subscription_until", "created_at", "updated_at",
	}).AddRow(
		m.ID, m.UserID, m.Slug, m.Title, m.Category, m.Description,
		m.Timezone, m.SubscriptionUntil, m.CreatedAt, m.UpdatedAt,
	)
}

func serviceRows(s domain.Service) *pgxmock.Rows {
	return pgxmock.NewRows([]string{
		"id", "master_id", "name", "description", "duration_minutes",
		"price", "is_active", "created_at", "updated_at",
	}).AddRow(
		s.ID, s.MasterID, s.Name, s.Description, s.DurationMinutes,
		s.Price, s.IsActive, s.CreatedAt, s.UpdatedAt,
	)
}

func scheduleRows(p domain.SchedulePeriod) *pgxmock.Rows {
	return pgxmock.NewRows([]string{
		"id", "master_id", "day_of_week", "start_time", "end_time",
		"created_at", "updated_at",
	}).AddRow(
		p.ID, p.MasterID, int16(p.DayOfWeek), p.StartTime, p.EndTime,
		p.CreatedAt, p.UpdatedAt,
	)
}

func TestNewMasterRepoRejectsNilDB(t *testing.T) {
	if _, err := NewMasterRepo(nil); err == nil {
		t.Fatal("NewMasterRepo(nil) expected an error")
	}
}

func TestCreateMaster(t *testing.T) {
	repo, mock := newMockRepo(t)
	input := masterTest
	input.ID = 0

	mock.ExpectQuery("INSERT INTO masters").
		WithArgs(input.UserID, input.Slug, input.Title, input.Category, input.Description,
			input.Timezone, input.SubscriptionUntil, input.CreatedAt, input.UpdatedAt).
		WillReturnRows(masterRows(masterTest))

	got, err := repo.CreateMaster(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateMaster() error = %v", err)
	}
	if got != masterTest {
		t.Errorf("CreateMaster() = %#v, want %#v", got, masterTest)
	}
	assertMockExpectations(t, mock)
}

func TestGetMasterByUserID(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery("FROM masters WHERE user_id").
		WithArgs(masterTest.UserID).
		WillReturnRows(masterRows(masterTest))

	got, err := repo.GetMasterByUserID(context.Background(), masterTest.UserID)
	if err != nil {
		t.Fatalf("GetMasterByUserID() error = %v", err)
	}
	if got != masterTest {
		t.Errorf("GetMasterByUserID() = %#v, want %#v", got, masterTest)
	}
	assertMockExpectations(t, mock)
}

func TestGetPublicMasterBySlug(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery("FROM masters WHERE slug").
		WithArgs(masterTest.Slug).
		WillReturnRows(masterRows(masterTest))

	got, err := repo.GetPublicMasterBySlug(context.Background(), masterTest.Slug)
	if err != nil {
		t.Fatalf("GetPublicMasterBySlug() error = %v", err)
	}
	if got != masterTest {
		t.Errorf("GetPublicMasterBySlug() = %#v, want %#v", got, masterTest)
	}
	assertMockExpectations(t, mock)
}

func TestUpdateMaster(t *testing.T) {
	repo, mock := newMockRepo(t)
	input := masterTest
	input.Title = "Updated title"
	input.UpdatedAt = masterTestTime.Add(time.Minute)

	mock.ExpectQuery("UPDATE masters").
		WithArgs(input.ID, input.UserID, input.Title, input.Category, input.Description,
			input.Timezone, input.UpdatedAt).
		WillReturnRows(masterRows(input))

	got, err := repo.UpdateMaster(context.Background(), input)
	if err != nil {
		t.Fatalf("UpdateMaster() error = %v", err)
	}
	if got != input {
		t.Errorf("UpdateMaster() = %#v, want %#v", got, input)
	}
	assertMockExpectations(t, mock)
}

func TestListServicesByMasterID(t *testing.T) {
	repo, mock := newMockRepo(t)
	second := serviceTest
	second.ID = 22
	second.Name = "Beard trim"

	mock.ExpectQuery("FROM services WHERE master_id").
		WithArgs(masterTest.ID).
		WillReturnRows(serviceRows(serviceTest).AddRow(
			second.ID, second.MasterID, second.Name, second.Description, second.DurationMinutes,
			second.Price, second.IsActive, second.CreatedAt, second.UpdatedAt,
		))

	got, err := repo.ListServicesByMasterID(context.Background(), masterTest.ID)
	if err != nil {
		t.Fatalf("ListServicesByMasterID() error = %v", err)
	}
	want := []domain.Service{serviceTest, second}
	if len(got) != len(want) {
		t.Fatalf("ListServicesByMasterID() returned %d services, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("service[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
	assertMockExpectations(t, mock)
}

func TestListActiveServicesByMasterIDFiltersActive(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery("is_active = TRUE").
		WithArgs(masterTest.ID).
		WillReturnRows(serviceRows(serviceTest))

	got, err := repo.ListActiveServicesByMasterID(context.Background(), masterTest.ID)
	if err != nil {
		t.Fatalf("ListActiveServicesByMasterID() error = %v", err)
	}
	if len(got) != 1 || got[0] != serviceTest {
		t.Errorf("ListActiveServicesByMasterID() = %#v, want [%#v]", got, serviceTest)
	}
	assertMockExpectations(t, mock)
}

func TestListServicesReturnsNonNilEmptySlice(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery("FROM services WHERE master_id").
		WithArgs(masterTest.ID).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "master_id", "name", "description", "duration_minutes",
			"price", "is_active", "created_at", "updated_at",
		}))

	got, err := repo.ListServicesByMasterID(context.Background(), masterTest.ID)
	if err != nil {
		t.Fatalf("ListServicesByMasterID() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("expected non-nil empty slice, got %#v", got)
	}
	assertMockExpectations(t, mock)
}

func TestCreateService(t *testing.T) {
	repo, mock := newMockRepo(t)
	input := serviceTest
	input.ID = 0

	mock.ExpectQuery("INSERT INTO services").
		WithArgs(input.MasterID, input.Name, input.Description, input.DurationMinutes,
			input.Price, input.IsActive, input.CreatedAt, input.UpdatedAt).
		WillReturnRows(serviceRows(serviceTest))

	got, err := repo.CreateService(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateService() error = %v", err)
	}
	if got != serviceTest {
		t.Errorf("CreateService() = %#v, want %#v", got, serviceTest)
	}
	assertMockExpectations(t, mock)
}

func TestGetServiceByIDAndMasterIDScopesByBothIDs(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery("WHERE id = \\$1 AND master_id = \\$2").
		WithArgs(serviceTest.ID, serviceTest.MasterID).
		WillReturnRows(serviceRows(serviceTest))

	got, err := repo.GetServiceByIDAndMasterID(context.Background(), serviceTest.ID, serviceTest.MasterID)
	if err != nil {
		t.Fatalf("GetServiceByIDAndMasterID() error = %v", err)
	}
	if got != serviceTest {
		t.Errorf("GetServiceByIDAndMasterID() = %#v, want %#v", got, serviceTest)
	}
	assertMockExpectations(t, mock)
}

func TestUpdateService(t *testing.T) {
	repo, mock := newMockRepo(t)
	input := serviceTest
	input.Name = "Updated haircut"
	input.IsActive = false
	input.UpdatedAt = masterTestTime.Add(time.Minute)

	mock.ExpectQuery("UPDATE services").
		WithArgs(input.ID, input.MasterID, input.Name, input.Description, input.DurationMinutes,
			input.Price, input.IsActive, input.UpdatedAt).
		WillReturnRows(serviceRows(input))

	got, err := repo.UpdateService(context.Background(), input)
	if err != nil {
		t.Fatalf("UpdateService() error = %v", err)
	}
	if got != input {
		t.Errorf("UpdateService() = %#v, want %#v", got, input)
	}
	assertMockExpectations(t, mock)
}

func TestReplaceSchedulePeriods(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM schedule_periods").
		WithArgs(scheduleTest.MasterID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectQuery("INSERT INTO schedule_periods").
		WithArgs(scheduleTest.MasterID, scheduleTest.DayOfWeek, scheduleTest.StartTime,
			scheduleTest.EndTime, scheduleTest.CreatedAt, scheduleTest.UpdatedAt).
		WillReturnRows(scheduleRows(scheduleTest))
	mock.ExpectCommit()

	got, err := repo.ReplaceSchedulePeriods(context.Background(), scheduleTest.MasterID, []domain.SchedulePeriod{scheduleTest})
	if err != nil {
		t.Fatalf("ReplaceSchedulePeriods() error = %v", err)
	}
	if len(got) != 1 || got[0] != scheduleTest {
		t.Errorf("ReplaceSchedulePeriods() = %#v, want [%#v]", got, scheduleTest)
	}
	assertMockExpectations(t, mock)
}

func TestReplaceSchedulePeriodsAllowsClearing(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM schedule_periods").
		WithArgs(masterTest.ID).
		WillReturnResult(pgxmock.NewResult("DELETE", 2))
	mock.ExpectCommit()

	got, err := repo.ReplaceSchedulePeriods(context.Background(), masterTest.ID, nil)
	if err != nil {
		t.Fatalf("ReplaceSchedulePeriods() error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("expected non-nil empty result, got %#v", got)
	}
	assertMockExpectations(t, mock)
}

func TestReplaceSchedulePeriodsRollsBackOnInsertError(t *testing.T) {
	repo, mock := newMockRepo(t)
	insertErr := errors.New("insert failed")
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM schedule_periods").
		WithArgs(scheduleTest.MasterID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectQuery("INSERT INTO schedule_periods").
		WithArgs(scheduleTest.MasterID, scheduleTest.DayOfWeek, scheduleTest.StartTime,
			scheduleTest.EndTime, scheduleTest.CreatedAt, scheduleTest.UpdatedAt).
		WillReturnError(insertErr)
	mock.ExpectRollback()

	_, err := repo.ReplaceSchedulePeriods(context.Background(), scheduleTest.MasterID, []domain.SchedulePeriod{scheduleTest})
	if !errors.Is(err, insertErr) {
		t.Errorf("ReplaceSchedulePeriods() error = %v, want wrapped %v", err, insertErr)
	}
	assertMockExpectations(t, mock)
}

func TestListSchedulePeriodsByMasterID(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery("FROM schedule_periods").
		WithArgs(scheduleTest.MasterID).
		WillReturnRows(scheduleRows(scheduleTest))

	got, err := repo.ListSchedulePeriodsByMasterID(context.Background(), scheduleTest.MasterID)
	if err != nil {
		t.Fatalf("ListSchedulePeriodsByMasterID() error = %v", err)
	}
	if len(got) != 1 || got[0] != scheduleTest {
		t.Errorf("ListSchedulePeriodsByMasterID() = %#v, want [%#v]", got, scheduleTest)
	}
	assertMockExpectations(t, mock)
}
