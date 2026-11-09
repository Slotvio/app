package service

import (
	"context"
	"errors"
	"testing"

	"github.com/aleksandrboro/slotvio/internal/domain"
)

var errUnexpectedRepositoryCall = errors.New("unexpected repository call")

type masterRepoStub struct {
	calls int

	createMasterFn                func(context.Context, domain.Master) (domain.Master, error)
	getMasterByUserIDFn           func(context.Context, int64) (domain.Master, error)
	getPublicMasterBySlugFn       func(context.Context, string) (domain.Master, error)
	updateMasterFn                func(context.Context, domain.Master) (domain.Master, error)
	listServicesByMasterIDFn      func(context.Context, int64) ([]domain.Service, error)
	listActiveServicesFn          func(context.Context, int64) ([]domain.Service, error)
	createServiceFn               func(context.Context, domain.Service) (domain.Service, error)
	getServiceByIDAndMasterIDFn   func(context.Context, int64, int64) (domain.Service, error)
	updateServiceFn               func(context.Context, domain.Service) (domain.Service, error)
	replaceSchedulePeriodsFn      func(context.Context, int64, []domain.SchedulePeriod) ([]domain.SchedulePeriod, error)
	listSchedulePeriodsByMasterFn func(context.Context, int64) ([]domain.SchedulePeriod, error)
}

func (r *masterRepoStub) CreateMaster(ctx context.Context, master domain.Master) (domain.Master, error) {
	r.calls++
	if r.createMasterFn == nil {
		return domain.Master{}, errUnexpectedRepositoryCall
	}
	return r.createMasterFn(ctx, master)
}

func (r *masterRepoStub) GetMasterByUserID(ctx context.Context, userID int64) (domain.Master, error) {
	r.calls++
	if r.getMasterByUserIDFn == nil {
		return domain.Master{}, errUnexpectedRepositoryCall
	}
	return r.getMasterByUserIDFn(ctx, userID)
}

func (r *masterRepoStub) GetPublicMasterBySlug(ctx context.Context, slug string) (domain.Master, error) {
	r.calls++
	if r.getPublicMasterBySlugFn == nil {
		return domain.Master{}, errUnexpectedRepositoryCall
	}
	return r.getPublicMasterBySlugFn(ctx, slug)
}

func (r *masterRepoStub) UpdateMaster(ctx context.Context, master domain.Master) (domain.Master, error) {
	r.calls++
	if r.updateMasterFn == nil {
		return domain.Master{}, errUnexpectedRepositoryCall
	}
	return r.updateMasterFn(ctx, master)
}

func (r *masterRepoStub) ListServicesByMasterID(ctx context.Context, masterID int64) ([]domain.Service, error) {
	r.calls++
	if r.listServicesByMasterIDFn == nil {
		return nil, errUnexpectedRepositoryCall
	}
	return r.listServicesByMasterIDFn(ctx, masterID)
}

func (r *masterRepoStub) ListActiveServicesByMasterID(ctx context.Context, masterID int64) ([]domain.Service, error) {
	r.calls++
	if r.listActiveServicesFn == nil {
		return nil, errUnexpectedRepositoryCall
	}
	return r.listActiveServicesFn(ctx, masterID)
}

func (r *masterRepoStub) CreateService(ctx context.Context, service domain.Service) (domain.Service, error) {
	r.calls++
	if r.createServiceFn == nil {
		return domain.Service{}, errUnexpectedRepositoryCall
	}
	return r.createServiceFn(ctx, service)
}

func (r *masterRepoStub) GetServiceByIDAndMasterID(ctx context.Context, serviceID, masterID int64) (domain.Service, error) {
	r.calls++
	if r.getServiceByIDAndMasterIDFn == nil {
		return domain.Service{}, errUnexpectedRepositoryCall
	}
	return r.getServiceByIDAndMasterIDFn(ctx, serviceID, masterID)
}

func (r *masterRepoStub) UpdateService(ctx context.Context, service domain.Service) (domain.Service, error) {
	r.calls++
	if r.updateServiceFn == nil {
		return domain.Service{}, errUnexpectedRepositoryCall
	}
	return r.updateServiceFn(ctx, service)
}

func (r *masterRepoStub) ReplaceSchedulePeriods(ctx context.Context, masterID int64, periods []domain.SchedulePeriod) ([]domain.SchedulePeriod, error) {
	r.calls++
	if r.replaceSchedulePeriodsFn == nil {
		return nil, errUnexpectedRepositoryCall
	}
	return r.replaceSchedulePeriodsFn(ctx, masterID, periods)
}

func (r *masterRepoStub) ListSchedulePeriodsByMasterID(ctx context.Context, masterID int64) ([]domain.SchedulePeriod, error) {
	r.calls++
	if r.listSchedulePeriodsByMasterFn == nil {
		return nil, errUnexpectedRepositoryCall
	}
	return r.listSchedulePeriodsByMasterFn(ctx, masterID)
}

func newMasterServiceForTest(t *testing.T, repo MasterRepository) *masterService {
	t.Helper()
	service, err := NewMasterService(nil, repo)
	if err != nil {
		t.Fatalf("NewMasterService() error = %v", err)
	}
	return service
}

func TestNewMasterServiceRequiresRepository(t *testing.T) {
	service, err := NewMasterService(nil, nil)
	if !errors.Is(err, ErrRepositoryNotConfigured) {
		t.Fatalf("error = %v, want ErrRepositoryNotConfigured", err)
	}
	if service != nil {
		t.Fatalf("service = %#v, want nil", service)
	}
}

func TestCreateProfile(t *testing.T) {
	var got domain.Master
	repo := &masterRepoStub{
		createMasterFn: func(_ context.Context, master domain.Master) (domain.Master, error) {
			got = master
			master.ID = 12
			return master, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	created, err := svc.CreateProfile(context.Background(), 7, CreateProfileInput{
		Slug:        "alex",
		Title:       "  Alex  ",
		Category:    "  barber ",
		Description: "  Profile description  ",
		Timezone:    " Europe/Minsk ",
	})
	if err != nil {
		t.Fatalf("CreateProfile() error = %v", err)
	}
	if created == nil || created.ID != 12 {
		t.Fatalf("created profile = %#v, want ID 12", created)
	}
	if got.UserID != 7 || got.Slug != "alex" || got.Title != "Alex" || got.Category != "barber" || got.Description != "Profile description" || got.Timezone != "Europe/Minsk" {
		t.Errorf("repository received master = %+v", got)
	}
	if got.CreatedAt.IsZero() || !got.CreatedAt.Equal(got.UpdatedAt) {
		t.Errorf("created_at/updated_at not set consistently: created=%v updated=%v", got.CreatedAt, got.UpdatedAt)
	}
	if want := got.CreatedAt.AddDate(0, trialPeriodMonths, 0); !got.SubscriptionUntil.Equal(want) {
		t.Errorf("subscription_until = %v, want %v", got.SubscriptionUntil, want)
	}
}

func TestCreateProfileRejectsInvalidInputBeforeRepository(t *testing.T) {
	repo := &masterRepoStub{}
	svc := newMasterServiceForTest(t, repo)

	tests := []struct {
		name   string
		userID int64
		input  CreateProfileInput
	}{
		{name: "invalid user id", userID: 0, input: CreateProfileInput{Slug: "alex", Title: "Alex", Category: "barber", Timezone: "Europe/Minsk"}},
		{name: "invalid slug", userID: 1, input: CreateProfileInput{Slug: "Alex!", Title: "Alex", Category: "barber", Timezone: "Europe/Minsk"}},
		{name: "blank title", userID: 1, input: CreateProfileInput{Slug: "alex", Title: "   ", Category: "barber", Timezone: "Europe/Minsk"}},
		{name: "invalid timezone", userID: 1, input: CreateProfileInput{Slug: "alex", Title: "Alex", Category: "barber", Timezone: "Nowhere/Invalid"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateProfile(context.Background(), tt.userID, tt.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
	if repo.calls != 0 {
		t.Errorf("repository calls = %d, want 0 for invalid inputs", repo.calls)
	}
}

func TestGetOwnProfile(t *testing.T) {
	repo := &masterRepoStub{
		getMasterByUserIDFn: func(_ context.Context, userID int64) (domain.Master, error) {
			if userID != 7 {
				t.Errorf("userID = %d, want 7", userID)
			}
			return domain.Master{ID: 3, UserID: userID, Slug: "alex"}, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	master, err := svc.GetOwnProfile(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetOwnProfile() error = %v", err)
	}
	if master == nil || master.ID != 3 || master.UserID != 7 {
		t.Errorf("profile = %#v", master)
	}
}

func TestGetPublicProfileIncludesOnlyRepositoryResultServices(t *testing.T) {
	repo := &masterRepoStub{
		getPublicMasterBySlugFn: func(_ context.Context, slug string) (domain.Master, error) {
			if slug != "alex" {
				t.Errorf("slug = %q, want alex", slug)
			}
			return domain.Master{ID: 3, Slug: slug}, nil
		},
		listActiveServicesFn: func(_ context.Context, masterID int64) ([]domain.Service, error) {
			if masterID != 3 {
				t.Errorf("masterID = %d, want 3", masterID)
			}
			return []domain.Service{{ID: 18, MasterID: masterID, Name: "Cut", IsActive: true}}, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	got, err := svc.GetPublicProfile(context.Background(), "alex")
	if err != nil {
		t.Fatalf("GetPublicProfile() error = %v", err)
	}
	if got == nil || got.Master.ID != 3 || len(got.Services) != 1 || got.Services[0].ID != 18 {
		t.Errorf("public profile output = %#v", got)
	}
}

func TestUpdateProfileAppliesOnlyProvidedFields(t *testing.T) {
	title := "  New title  "
	emptyDescription := ""
	var updated domain.Master
	repo := &masterRepoStub{
		getMasterByUserIDFn: func(_ context.Context, userID int64) (domain.Master, error) {
			return domain.Master{ID: 3, UserID: userID, Slug: "alex", Title: "Old", Category: "barber", Description: "Old description", Timezone: "Europe/Minsk"}, nil
		},
		updateMasterFn: func(_ context.Context, master domain.Master) (domain.Master, error) {
			updated = master
			return master, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	got, err := svc.UpdateProfile(context.Background(), 7, UpdateProfileInput{Title: &title, Description: &emptyDescription})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got == nil || got.Title != "New title" || got.Description != "" || got.Category != "barber" || got.Timezone != "Europe/Minsk" {
		t.Errorf("updated profile = %#v", got)
	}
	if updated.UpdatedAt.IsZero() {
		t.Error("UpdatedAt was not set")
	}
}

func TestListServicesUsesMasterID(t *testing.T) {
	repo := &masterRepoStub{
		getMasterByUserIDFn: func(_ context.Context, userID int64) (domain.Master, error) {
			return domain.Master{ID: 3, UserID: userID}, nil
		},
		listServicesByMasterIDFn: func(_ context.Context, masterID int64) ([]domain.Service, error) {
			if masterID != 3 {
				t.Errorf("masterID = %d, want 3", masterID)
			}
			return []domain.Service{{ID: 18, MasterID: masterID}}, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	services, err := svc.ListServices(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListServices() error = %v", err)
	}
	if len(services) != 1 || services[0].ID != 18 {
		t.Errorf("services = %+v", services)
	}
}

func TestCreateServiceSetsOwnerAndDefaults(t *testing.T) {
	var got domain.Service
	repo := &masterRepoStub{
		getMasterByUserIDFn: func(_ context.Context, userID int64) (domain.Master, error) {
			return domain.Master{ID: 3, UserID: userID}, nil
		},
		createServiceFn: func(_ context.Context, service domain.Service) (domain.Service, error) {
			got = service
			service.ID = 18
			return service, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	created, err := svc.CreateService(context.Background(), 7, CreateServiceInput{
		Name: "  Haircut  ", Description: "  Short desc  ", DurationMinutes: 45, Price: 5000,
	})
	if err != nil {
		t.Fatalf("CreateService() error = %v", err)
	}
	if created == nil || created.ID != 18 {
		t.Fatalf("created service = %#v", created)
	}
	if got.MasterID != 3 || got.Name != "Haircut" || got.Description != "Short desc" || got.DurationMinutes != 45 || got.Price != 5000 || !got.IsActive {
		t.Errorf("repository received service = %+v", got)
	}
	if got.CreatedAt.IsZero() || !got.CreatedAt.Equal(got.UpdatedAt) {
		t.Errorf("service timestamps not set consistently: %+v", got)
	}
}

func TestUpdateServiceScopesLookupToMasterAndAppliesFalse(t *testing.T) {
	isActive := false
	var lookedUpServiceID, lookedUpMasterID int64
	var updated domain.Service
	repo := &masterRepoStub{
		getMasterByUserIDFn: func(_ context.Context, userID int64) (domain.Master, error) {
			return domain.Master{ID: 3, UserID: userID}, nil
		},
		getServiceByIDAndMasterIDFn: func(_ context.Context, serviceID, masterID int64) (domain.Service, error) {
			lookedUpServiceID, lookedUpMasterID = serviceID, masterID
			return domain.Service{ID: serviceID, MasterID: masterID, Name: "Haircut", IsActive: true}, nil
		},
		updateServiceFn: func(_ context.Context, service domain.Service) (domain.Service, error) {
			updated = service
			return service, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	got, err := svc.UpdateService(context.Background(), 7, 18, UpdateServiceInput{IsActive: &isActive})
	if err != nil {
		t.Fatalf("UpdateService() error = %v", err)
	}
	if lookedUpServiceID != 18 || lookedUpMasterID != 3 {
		t.Errorf("lookup used serviceID=%d masterID=%d", lookedUpServiceID, lookedUpMasterID)
	}
	if got == nil || got.IsActive || updated.IsActive || updated.UpdatedAt.IsZero() {
		t.Errorf("updated service = got:%#v stored:%+v", got, updated)
	}
}

func TestReplaceScheduleParsesClockTimes(t *testing.T) {
	var gotMasterID int64
	var gotPeriods []domain.SchedulePeriod
	repo := &masterRepoStub{
		getMasterByUserIDFn: func(_ context.Context, userID int64) (domain.Master, error) {
			return domain.Master{ID: 3, UserID: userID}, nil
		},
		replaceSchedulePeriodsFn: func(_ context.Context, masterID int64, periods []domain.SchedulePeriod) ([]domain.SchedulePeriod, error) {
			gotMasterID, gotPeriods = masterID, periods
			return periods, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	got, err := svc.ReplaceSchedule(context.Background(), 7, ReplaceScheduleInput{Periods: []SchedulePeriodInput{{
		DayOfWeek: 1, StartTime: "09:00", EndTime: "17:30",
	}}})
	if err != nil {
		t.Fatalf("ReplaceSchedule() error = %v", err)
	}
	if gotMasterID != 3 || len(gotPeriods) != 1 || gotPeriods[0].MasterID != 3 {
		t.Fatalf("replacement args: masterID=%d periods=%+v", gotMasterID, gotPeriods)
	}
	if gotPeriods[0].StartTime.Format("15:04") != "09:00" || gotPeriods[0].EndTime.Format("15:04") != "17:30" {
		t.Errorf("stored times = %s-%s", gotPeriods[0].StartTime.Format("15:04"), gotPeriods[0].EndTime.Format("15:04"))
	}
	if len(got) != 1 || got[0].DayOfWeek != 1 {
		t.Errorf("returned schedule = %+v", got)
	}
}

func TestReplaceScheduleAllowsClearingAllPeriods(t *testing.T) {
	var gotPeriods []domain.SchedulePeriod
	repo := &masterRepoStub{
		getMasterByUserIDFn: func(context.Context, int64) (domain.Master, error) {
			return domain.Master{ID: 3}, nil
		},
		replaceSchedulePeriodsFn: func(_ context.Context, _ int64, periods []domain.SchedulePeriod) ([]domain.SchedulePeriod, error) {
			gotPeriods = periods
			return periods, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	got, err := svc.ReplaceSchedule(context.Background(), 7, ReplaceScheduleInput{})
	if err != nil {
		t.Fatalf("ReplaceSchedule(empty) error = %v", err)
	}
	if gotPeriods == nil || len(gotPeriods) != 0 || got == nil || len(got) != 0 {
		t.Errorf("expected empty non-nil schedule slices; stored=%#v returned=%#v", gotPeriods, got)
	}
}

func TestGetScheduleUsesMasterID(t *testing.T) {
	repo := &masterRepoStub{
		getMasterByUserIDFn: func(_ context.Context, userID int64) (domain.Master, error) {
			return domain.Master{ID: 3, UserID: userID}, nil
		},
		listSchedulePeriodsByMasterFn: func(_ context.Context, masterID int64) ([]domain.SchedulePeriod, error) {
			if masterID != 3 {
				t.Errorf("masterID = %d, want 3", masterID)
			}
			return []domain.SchedulePeriod{{MasterID: masterID, DayOfWeek: 1}}, nil
		},
	}
	svc := newMasterServiceForTest(t, repo)

	schedule, err := svc.GetSchedule(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetSchedule() error = %v", err)
	}
	if len(schedule) != 1 || schedule[0].MasterID != 3 {
		t.Errorf("schedule = %+v", schedule)
	}
}

func TestServiceValidationRunsBeforeRepository(t *testing.T) {
	repo := &masterRepoStub{}
	svc := newMasterServiceForTest(t, repo)
	ctx := context.Background()
	validProfile := CreateProfileInput{Slug: "alex", Title: "Alex", Category: "barber", Timezone: "Europe/Minsk"}
	blank := "   "
	badDuration := 1

	tests := []struct {
		name string
		call func() error
	}{
		{name: "get own profile user id", call: func() error { _, err := svc.GetOwnProfile(ctx, 0); return err }},
		{name: "public profile slug", call: func() error { _, err := svc.GetPublicProfile(ctx, "BAD"); return err }},
		{name: "update profile empty patch", call: func() error { _, err := svc.UpdateProfile(ctx, 7, UpdateProfileInput{}); return err }},
		{name: "list services user id", call: func() error { _, err := svc.ListServices(ctx, 0); return err }},
		{name: "create service duration", call: func() error {
			_, err := svc.CreateService(ctx, 7, CreateServiceInput{Name: "Cut", DurationMinutes: 1})
			return err
		}},
		{name: "update service empty patch", call: func() error { _, err := svc.UpdateService(ctx, 7, 5, UpdateServiceInput{}); return err }},
		{name: "update service duration", call: func() error {
			_, err := svc.UpdateService(ctx, 7, 5, UpdateServiceInput{DurationMinutes: &badDuration})
			return err
		}},
		{name: "replace schedule invalid day", call: func() error {
			_, err := svc.ReplaceSchedule(ctx, 7, ReplaceScheduleInput{Periods: []SchedulePeriodInput{{DayOfWeek: 0, StartTime: "09:00", EndTime: "10:00"}}})
			return err
		}},
		{name: "get schedule user id", call: func() error { _, err := svc.GetSchedule(ctx, 0); return err }},
		{name: "create profile blank title", call: func() error {
			input := validProfile
			input.Title = blank
			_, err := svc.CreateProfile(ctx, 7, input)
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
	if repo.calls != 0 {
		t.Errorf("repository calls = %d, want 0 when validation fails", repo.calls)
	}
}

func TestRepositoryErrorsAreWrapped(t *testing.T) {
	repositoryErr := errors.New("db unavailable")
	repo := &masterRepoStub{
		createMasterFn: func(context.Context, domain.Master) (domain.Master, error) {
			return domain.Master{}, repositoryErr
		},
	}
	svc := newMasterServiceForTest(t, repo)

	_, err := svc.CreateProfile(context.Background(), 7, CreateProfileInput{
		Slug: "alex", Title: "Alex", Category: "barber", Timezone: "Europe/Minsk",
	})
	if !errors.Is(err, repositoryErr) {
		t.Fatalf("error = %v, want wrapped repository error", err)
	}
}
