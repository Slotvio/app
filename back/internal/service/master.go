package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aleksandrboro/slotvio/internal/domain"
)

var ErrRepositoryNotConfigured = errors.New("master repository is not configured")

const trialPeriodMonths = 1

type MasterRepository interface {
	CreateMaster(ctx context.Context, master domain.Master) (domain.Master, error)
	GetMasterByUserID(ctx context.Context, userID int64) (domain.Master, error)
	GetPublicMasterBySlug(ctx context.Context, slug string) (domain.Master, error)
	UpdateMaster(ctx context.Context, master domain.Master) (domain.Master, error)

	ListServicesByMasterID(ctx context.Context, masterID int64) ([]domain.Service, error)
	ListActiveServicesByMasterID(ctx context.Context, masterID int64) ([]domain.Service, error)
	CreateService(ctx context.Context, service domain.Service) (domain.Service, error)
	GetServiceByIDAndMasterID(ctx context.Context, serviceID, masterID int64) (domain.Service, error)
	UpdateService(ctx context.Context, service domain.Service) (domain.Service, error)

	ReplaceSchedulePeriods(ctx context.Context, masterID int64, periods []domain.SchedulePeriod) ([]domain.SchedulePeriod, error)
	ListSchedulePeriodsByMasterID(ctx context.Context, masterID int64) ([]domain.SchedulePeriod, error)
}

type masterService struct {
	log        *slog.Logger
	masterRepo MasterRepository
}

func NewMasterService(log *slog.Logger, masterRepo MasterRepository) (*masterService, error) {
	if masterRepo == nil {
		return nil, ErrRepositoryNotConfigured
	}
	if log == nil {
		log = slog.Default()
	}

	return &masterService{
		log:        log,
		masterRepo: masterRepo,
	}, nil
}

func (s *masterService) masterByUserID(ctx context.Context, userID int64) (domain.Master, error) {
	master, err := s.masterRepo.GetMasterByUserID(ctx, userID)
	if err != nil {
		return domain.Master{}, fmt.Errorf("get master by user id: %w", err)
	}
	return master, nil
}

func (s *masterService) CreateProfile(ctx context.Context, userID int64, input CreateProfileInput) (*domain.Master, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Category = strings.TrimSpace(input.Category)
	input.Description = strings.TrimSpace(input.Description)
	input.Timezone = strings.TrimSpace(input.Timezone)

	if err := validateCreateProfileInput(userID, input); err != nil {
		return nil, fmt.Errorf("validate create profile: %w", err)
	}

	now := time.Now().UTC()
	masterData := domain.Master{
		UserID:            userID,
		Slug:              input.Slug,
		Title:             input.Title,
		Category:          input.Category,
		Description:       input.Description,
		Timezone:          input.Timezone,
		SubscriptionUntil: now.AddDate(0, trialPeriodMonths, 0),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	master, err := s.masterRepo.CreateMaster(ctx, masterData)
	if err != nil {
		return nil, fmt.Errorf("create master: %w", err)
	}
	return &master, nil
}

func (s *masterService) GetOwnProfile(ctx context.Context, userID int64) (*domain.Master, error) {
	if err := validateGetOwnProfileInput(userID); err != nil {
		return nil, fmt.Errorf("validate get own profile: %w", err)
	}

	master, err := s.masterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get own profile: %w", err)
	}
	return &master, nil
}

func (s *masterService) GetPublicProfile(ctx context.Context, slug string) (*PublicProfileOutput, error) {
	if err := validateGetPublicProfileInput(slug); err != nil {
		return nil, fmt.Errorf("validate get public profile: %w", err)
	}

	master, err := s.masterRepo.GetPublicMasterBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get public master by slug: %w", err)
	}

	services, err := s.masterRepo.ListActiveServicesByMasterID(ctx, master.ID)
	if err != nil {
		return nil, fmt.Errorf("list active services for public profile: %w", err)
	}

	return &PublicProfileOutput{Master: master, Services: services}, nil
}

func (s *masterService) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*domain.Master, error) {
	input.Title = trimOptionalString(input.Title)
	input.Category = trimOptionalString(input.Category)
	input.Description = trimOptionalString(input.Description)
	input.Timezone = trimOptionalString(input.Timezone)

	if err := validateUpdateProfileInput(userID, input); err != nil {
		return nil, fmt.Errorf("validate update profile: %w", err)
	}

	master, err := s.masterRepo.GetMasterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get master for profile update: %w", err)
	}

	if input.Title != nil {
		master.Title = *input.Title
	}
	if input.Category != nil {
		master.Category = *input.Category
	}
	if input.Description != nil {
		master.Description = *input.Description
	}
	if input.Timezone != nil {
		master.Timezone = *input.Timezone
	}
	master.UpdatedAt = time.Now().UTC()

	updated, err := s.masterRepo.UpdateMaster(ctx, master)
	if err != nil {
		return nil, fmt.Errorf("update master profile: %w", err)
	}
	return &updated, nil
}

func (s *masterService) ListServices(ctx context.Context, userID int64) ([]domain.Service, error) {
	if err := validateListServicesInput(userID); err != nil {
		return nil, fmt.Errorf("validate list services: %w", err)
	}

	master, err := s.masterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get master for services: %w", err)
	}

	services, err := s.masterRepo.ListServicesByMasterID(ctx, master.ID)
	if err != nil {
		return nil, fmt.Errorf("list services by master id: %w", err)
	}
	return services, nil
}

func (s *masterService) CreateService(ctx context.Context, userID int64, input CreateServiceInput) (*domain.Service, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if err := validateCreateServiceInput(userID, input); err != nil {
		return nil, fmt.Errorf("validate create service: %w", err)
	}

	master, err := s.masterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get master for service creation: %w", err)
	}

	now := time.Now().UTC()
	serviceData := domain.Service{
		MasterID:        master.ID,
		Name:            input.Name,
		Description:     input.Description,
		DurationMinutes: input.DurationMinutes,
		Price:           input.Price,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	created, err := s.masterRepo.CreateService(ctx, serviceData)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}
	return &created, nil
}

func (s *masterService) UpdateService(ctx context.Context, userID int64, serviceID int64, input UpdateServiceInput) (*domain.Service, error) {
	input.Name = trimOptionalString(input.Name)
	input.Description = trimOptionalString(input.Description)
	if err := validateUpdateServiceInput(userID, serviceID, input); err != nil {
		return nil, fmt.Errorf("validate update service: %w", err)
	}

	master, err := s.masterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get master for service update: %w", err)
	}

	serviceData, err := s.masterRepo.GetServiceByIDAndMasterID(ctx, serviceID, master.ID)
	if err != nil {
		return nil, fmt.Errorf("get service by id and master id: %w", err)
	}

	if input.Name != nil {
		serviceData.Name = *input.Name
	}
	if input.Description != nil {
		serviceData.Description = *input.Description
	}
	if input.DurationMinutes != nil {
		serviceData.DurationMinutes = *input.DurationMinutes
	}
	if input.Price != nil {
		serviceData.Price = *input.Price
	}
	if input.IsActive != nil {
		serviceData.IsActive = *input.IsActive
	}
	serviceData.UpdatedAt = time.Now().UTC()

	updated, err := s.masterRepo.UpdateService(ctx, serviceData)
	if err != nil {
		return nil, fmt.Errorf("update service: %w", err)
	}
	return &updated, nil
}

func (s *masterService) ReplaceSchedule(ctx context.Context, userID int64, input ReplaceScheduleInput) ([]domain.SchedulePeriod, error) {
	if err := validateReplaceScheduleInput(userID, input); err != nil {
		return nil, fmt.Errorf("validate replace schedule: %w", err)
	}

	master, err := s.masterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get master for schedule replacement: %w", err)
	}

	now := time.Now().UTC()
	periods := make([]domain.SchedulePeriod, 0, len(input.Periods))
	for _, inputPeriod := range input.Periods {
		start, err := time.Parse("15:04", inputPeriod.StartTime)
		if err != nil {
			return nil, fmt.Errorf("parse schedule start_time: %w", err)
		}
		end, err := time.Parse("15:04", inputPeriod.EndTime)
		if err != nil {
			return nil, fmt.Errorf("parse schedule end_time: %w", err)
		}

		periods = append(periods, domain.SchedulePeriod{
			MasterID:  master.ID,
			DayOfWeek: inputPeriod.DayOfWeek,
			StartTime: start,
			EndTime:   end,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	saved, err := s.masterRepo.ReplaceSchedulePeriods(ctx, master.ID, periods)
	if err != nil {
		return nil, fmt.Errorf("replace schedule periods: %w", err)
	}
	return saved, nil
}

func (s *masterService) GetSchedule(ctx context.Context, userID int64) ([]domain.SchedulePeriod, error) {
	if err := validateGetScheduleInput(userID); err != nil {
		return nil, fmt.Errorf("validate get schedule: %w", err)
	}

	master, err := s.masterByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get master for schedule: %w", err)
	}

	repo := s.masterRepo
	schedule, err := repo.ListSchedulePeriodsByMasterID(ctx, master.ID)
	if err != nil {
		return nil, fmt.Errorf("list schedule periods by master id: %w", err)
	}
	return schedule, nil
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}
