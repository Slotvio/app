package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aleksandrboro/slotvio/internal/domain"
	requestid "github.com/aleksandrboro/slotvio/internal/request_id"
	"github.com/aleksandrboro/slotvio/internal/service"
	"github.com/gorilla/mux"
)

var errUnexpectedServiceCall = errors.New("unexpected service call")

type masterServiceStub struct {
	createProfileFn   func(context.Context, int64, service.CreateProfileInput) (*domain.Master, error)
	getOwnProfileFn   func(context.Context, int64) (*domain.Master, error)
	getPublicFn       func(context.Context, string) (*service.PublicProfileOutput, error)
	updateProfileFn   func(context.Context, int64, service.UpdateProfileInput) (*domain.Master, error)
	listServicesFn    func(context.Context, int64) ([]domain.Service, error)
	createServiceFn   func(context.Context, int64, service.CreateServiceInput) (*domain.Service, error)
	updateServiceFn   func(context.Context, int64, int64, service.UpdateServiceInput) (*domain.Service, error)
	replaceScheduleFn func(context.Context, int64, service.ReplaceScheduleInput) ([]domain.SchedulePeriod, error)
	getScheduleFn     func(context.Context, int64) ([]domain.SchedulePeriod, error)
}

func (s *masterServiceStub) CreateProfile(ctx context.Context, userID int64, input service.CreateProfileInput) (*domain.Master, error) {
	if s.createProfileFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.createProfileFn(ctx, userID, input)
}

func (s *masterServiceStub) GetOwnProfile(ctx context.Context, userID int64) (*domain.Master, error) {
	if s.getOwnProfileFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.getOwnProfileFn(ctx, userID)
}

func (s *masterServiceStub) GetPublicProfile(ctx context.Context, slug string) (*service.PublicProfileOutput, error) {
	if s.getPublicFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.getPublicFn(ctx, slug)
}

func (s *masterServiceStub) UpdateProfile(ctx context.Context, userID int64, input service.UpdateProfileInput) (*domain.Master, error) {
	if s.updateProfileFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.updateProfileFn(ctx, userID, input)
}

func (s *masterServiceStub) ListServices(ctx context.Context, userID int64) ([]domain.Service, error) {
	if s.listServicesFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.listServicesFn(ctx, userID)
}

func (s *masterServiceStub) CreateService(ctx context.Context, userID int64, input service.CreateServiceInput) (*domain.Service, error) {
	if s.createServiceFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.createServiceFn(ctx, userID, input)
}

func (s *masterServiceStub) UpdateService(ctx context.Context, userID, serviceID int64, input service.UpdateServiceInput) (*domain.Service, error) {
	if s.updateServiceFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.updateServiceFn(ctx, userID, serviceID, input)
}

func (s *masterServiceStub) ReplaceSchedule(ctx context.Context, userID int64, input service.ReplaceScheduleInput) ([]domain.SchedulePeriod, error) {
	if s.replaceScheduleFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.replaceScheduleFn(ctx, userID, input)
}

func (s *masterServiceStub) GetSchedule(ctx context.Context, userID int64) ([]domain.SchedulePeriod, error) {
	if s.getScheduleFn == nil {
		return nil, errUnexpectedServiceCall
	}
	return s.getScheduleFn(ctx, userID)
}

func newHandlerForTest(svc MasterService) *MasterHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewMasterHandler(logger, svc)
}

func requestWithID(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	return req.WithContext(requestid.WithRequestID(req.Context(), "test-request-id"))
}

func TestHandleCreateProfile(t *testing.T) {
	var gotUserID int64
	var gotInput service.CreateProfileInput
	stub := &masterServiceStub{
		createProfileFn: func(_ context.Context, userID int64, input service.CreateProfileInput) (*domain.Master, error) {
			gotUserID, gotInput = userID, input
			return &domain.Master{
				ID:          7,
				Slug:        input.Slug,
				Title:       input.Title,
				Category:    input.Category,
				Description: input.Description,
				Timezone:    input.Timezone,
			}, nil
		},
	}

	req := requestWithID(http.MethodPost, "/api/v1/master/profile", `{"slug":"alex","title":"Alex","category":"barber","description":"Profile","timezone":"Europe/Minsk"}`)
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleCreateProfile(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if gotUserID != 1 {
		t.Errorf("userID = %d, want 1 (current auth test stub)", gotUserID)
	}
	if gotInput.Slug != "alex" || gotInput.Timezone != "Europe/Minsk" {
		t.Errorf("service input = %+v, want slug alex and timezone Europe/Minsk", gotInput)
	}

	var got struct {
		ID   int64  `json:"id"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != 7 || got.Slug != "alex" {
		t.Errorf("response = %+v, want ID 7 and slug alex", got)
	}
}

func TestHandleCreateProfileRejectsInvalidInput(t *testing.T) {
	calls := 0
	stub := &masterServiceStub{
		createProfileFn: func(context.Context, int64, service.CreateProfileInput) (*domain.Master, error) {
			calls++
			return nil, nil
		},
	}
	h := newHandlerForTest(stub)

	tests := []struct {
		name string
		body string
	}{
		{name: "invalid slug", body: `{"slug":"Bad Slug","title":"Alex","category":"barber","timezone":"Europe/Minsk"}`},
		{name: "invalid timezone", body: `{"slug":"alex","title":"Alex","category":"barber","timezone":"Mars/Olympus"}`},
		{name: "malformed JSON", body: `{"slug":`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := requestWithID(http.MethodPost, "/api/v1/master/profile", tt.body)
			rec := httptest.NewRecorder()
			h.HandleCreateProfile(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
	if calls != 0 {
		t.Errorf("CreateProfile called %d times for invalid requests, want 0", calls)
	}
}

func TestHandleOwnProfile(t *testing.T) {
	stub := &masterServiceStub{
		getOwnProfileFn: func(_ context.Context, userID int64) (*domain.Master, error) {
			if userID != 1 {
				t.Errorf("userID = %d, want 1", userID)
			}
			return &domain.Master{Slug: "alex", Title: "Alex", Timezone: "Europe/Minsk"}, nil
		},
	}
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleOwnProfile(rec, requestWithID(http.MethodGet, "/api/v1/master/profile", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got struct {
		Slug     string `json:"slug"`
		Title    string `json:"title"`
		Timezone string `json:"timezone"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Slug != "alex" || got.Title != "Alex" || got.Timezone != "Europe/Minsk" {
		t.Errorf("response = %+v", got)
	}
}

func TestHandlePublicProfileIncludesServiceIDs(t *testing.T) {
	stub := &masterServiceStub{
		getPublicFn: func(_ context.Context, slug string) (*service.PublicProfileOutput, error) {
			if slug != "alex" {
				t.Errorf("slug = %q, want alex", slug)
			}
			return &service.PublicProfileOutput{
				Master: domain.Master{Slug: "alex", Title: "Alex", Timezone: "Europe/Minsk"},
				Services: []domain.Service{{
					ID: 19, Name: "Haircut", DurationMinutes: 45, Price: 5000, IsActive: true,
				}},
			}, nil
		},
	}
	req := requestWithID(http.MethodGet, "/api/v1/masters/alex", "")
	req = mux.SetURLVars(req, map[string]string{"slug": "alex"})
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandlePublicProfile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got struct {
		Services []struct {
			ID int64 `json:"id"`
		} `json:"services"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Services) != 1 || got.Services[0].ID != 19 {
		t.Errorf("services = %+v, want one service with ID 19", got.Services)
	}
}

func TestHandleUpdateProfile(t *testing.T) {
	var gotUserID int64
	var gotInput service.UpdateProfileInput
	title := "Updated"
	stub := &masterServiceStub{
		updateProfileFn: func(_ context.Context, userID int64, input service.UpdateProfileInput) (*domain.Master, error) {
			gotUserID, gotInput = userID, input
			return &domain.Master{ID: 3, Slug: "alex", Title: *input.Title}, nil
		},
	}
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleUpdateProfile(rec, requestWithID(http.MethodPatch, "/api/v1/master/profile", `{"title":"Updated"}`))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if gotUserID != 1 || gotInput.Title == nil || *gotInput.Title != title {
		t.Errorf("service call: userID=%d input=%+v", gotUserID, gotInput)
	}
}

func TestHandleUpdateProfileRejectsEmptyPatch(t *testing.T) {
	called := false
	stub := &masterServiceStub{
		updateProfileFn: func(context.Context, int64, service.UpdateProfileInput) (*domain.Master, error) {
			called = true
			return &domain.Master{}, nil
		},
	}
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleUpdateProfile(rec, requestWithID(http.MethodPatch, "/api/v1/master/profile", `{}`))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if called {
		t.Error("UpdateProfile should not be called for an empty PATCH")
	}
}

func TestHandleListServicesReturnsEmptyArray(t *testing.T) {
	stub := &masterServiceStub{
		listServicesFn: func(_ context.Context, userID int64) ([]domain.Service, error) {
			if userID != 1 {
				t.Errorf("userID = %d, want 1", userID)
			}
			return nil, nil
		},
	}
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleListServices(rec, requestWithID(http.MethodGet, "/api/v1/master/services", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got struct {
		Services []json.RawMessage `json:"services"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Services == nil || len(got.Services) != 0 {
		t.Errorf("services = %#v, want a non-nil empty array", got.Services)
	}
}

func TestHandleCreateServiceReturnsCreated(t *testing.T) {
	var gotInput service.CreateServiceInput
	stub := &masterServiceStub{
		createServiceFn: func(_ context.Context, userID int64, input service.CreateServiceInput) (*domain.Service, error) {
			if userID != 1 {
				t.Errorf("userID = %d, want 1", userID)
			}
			gotInput = input
			return &domain.Service{ID: 42, Name: input.Name, DurationMinutes: input.DurationMinutes, Price: input.Price, IsActive: true}, nil
		},
	}
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleCreateService(rec, requestWithID(http.MethodPost, "/api/v1/master/services", `{"name":"Haircut","duration_minutes":45,"price":5000}`))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if gotInput.Name != "Haircut" || gotInput.DurationMinutes != 45 || gotInput.Price != 5000 {
		t.Errorf("service input = %+v", gotInput)
	}
	var got struct {
		ID       int64 `json:"id"`
		IsActive bool  `json:"is_active"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != 42 || !got.IsActive {
		t.Errorf("response = %+v, want ID 42 and active service", got)
	}
}

func TestHandleCreateServiceNilResultDoesNotPanic(t *testing.T) {
	stub := &masterServiceStub{
		createServiceFn: func(context.Context, int64, service.CreateServiceInput) (*domain.Service, error) {
			return nil, nil
		},
	}
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleCreateService(rec, requestWithID(http.MethodPost, "/api/v1/master/services", `{"name":"Haircut","duration_minutes":45,"price":5000}`))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

func TestHandleUpdateService(t *testing.T) {
	var gotUserID, gotServiceID int64
	var gotInput service.UpdateServiceInput
	stub := &masterServiceStub{
		updateServiceFn: func(_ context.Context, userID, serviceID int64, input service.UpdateServiceInput) (*domain.Service, error) {
			gotUserID, gotServiceID, gotInput = userID, serviceID, input
			return &domain.Service{ID: serviceID, Name: "Haircut", IsActive: false}, nil
		},
	}
	req := requestWithID(http.MethodPatch, "/api/v1/master/services/42", `{"is_active":false}`)
	req = mux.SetURLVars(req, map[string]string{"id": "42"})
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleUpdateService(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if gotUserID != 1 || gotServiceID != 42 || gotInput.IsActive == nil || *gotInput.IsActive {
		t.Errorf("service call: userID=%d serviceID=%d input=%+v", gotUserID, gotServiceID, gotInput)
	}
}

func TestHandleUpdateServiceRejectsEmptyPatch(t *testing.T) {
	called := false
	stub := &masterServiceStub{
		updateServiceFn: func(context.Context, int64, int64, service.UpdateServiceInput) (*domain.Service, error) {
			called = true
			return &domain.Service{}, nil
		},
	}
	req := requestWithID(http.MethodPatch, "/api/v1/master/services/42", `{}`)
	req = mux.SetURLVars(req, map[string]string{"id": "42"})
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleUpdateService(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if called {
		t.Error("UpdateService should not be called for an empty PATCH")
	}
}

func TestHandleReplaceSchedule(t *testing.T) {
	var gotUserID int64
	var gotInput service.ReplaceScheduleInput
	stub := &masterServiceStub{
		replaceScheduleFn: func(_ context.Context, userID int64, input service.ReplaceScheduleInput) ([]domain.SchedulePeriod, error) {
			gotUserID, gotInput = userID, input
			return []domain.SchedulePeriod{{
				DayOfWeek: 1,
				StartTime: time.Date(2000, 1, 1, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2000, 1, 1, 17, 0, 0, 0, time.UTC),
			}}, nil
		},
	}
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleReplaceSchedule(rec, requestWithID(http.MethodPut, "/api/v1/master/schedule", `{"days":[{"day_of_week":1,"start_time":"09:00","end_time":"17:00"}]}`))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if gotUserID != 1 || len(gotInput.Periods) != 1 || gotInput.Periods[0].StartTime != "09:00" || gotInput.Periods[0].EndTime != "17:00" {
		t.Errorf("service call: userID=%d input=%+v", gotUserID, gotInput)
	}
	var got struct {
		Days []struct {
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		} `json:"days"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Days) != 1 || got.Days[0].StartTime != "09:00" || got.Days[0].EndTime != "17:00" {
		t.Errorf("schedule response = %+v", got)
	}
}

func TestHandleGetSchedule(t *testing.T) {
	stub := &masterServiceStub{
		getScheduleFn: func(_ context.Context, userID int64) ([]domain.SchedulePeriod, error) {
			if userID != 1 {
				t.Errorf("userID = %d, want 1", userID)
			}
			return []domain.SchedulePeriod{{
				DayOfWeek: 2,
				StartTime: time.Date(2000, 1, 1, 10, 30, 0, 0, time.UTC),
				EndTime:   time.Date(2000, 1, 1, 18, 0, 0, 0, time.UTC),
			}}, nil
		},
	}
	rec := httptest.NewRecorder()
	newHandlerForTest(stub).HandleGetSchedule(rec, requestWithID(http.MethodGet, "/api/v1/master/schedule", ""))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got struct {
		Days []struct {
			DayOfWeek int8   `json:"day_of_week"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		} `json:"days"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got.Days) != 1 || got.Days[0].DayOfWeek != 2 || got.Days[0].StartTime != "10:30" || got.Days[0].EndTime != "18:00" {
		t.Errorf("schedule response = %+v", got)
	}
}
