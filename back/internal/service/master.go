package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aleksandrboro/slotvio/internal/domain"
)

var (
	errUnimplemented = errors.New("UNIMPLEMENTED!!!")
)

type masterService struct {
	log        *slog.Logger
	masterRepo MasterRepository
}

func NewMasterService(log *slog.Logger, masterRepo MasterRepository) *masterService {
	return &masterService{
		log:        log,
		masterRepo: masterRepo,
	}
}

type MasterRepository interface{}

func (s *masterService) CreateProfile(ctx context.Context, userID int64, input CreateProfileInput) (*domain.Master, error) {
	return nil, errUnimplemented
}

func (s *masterService) GetOwnProfile(ctx context.Context, userID int64) (*domain.Master, error) {
	return nil, errUnimplemented
}

func (s *masterService) GetPublicProfile(ctx context.Context, slug string) (*PublicProfileOutput, error) {
	return nil, errUnimplemented
}

func (s *masterService) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*domain.Master, error) {
	return nil, errUnimplemented
}

func (s *masterService) ListServices(ctx context.Context, userID int64) ([]domain.Service, error) {
	return nil, errUnimplemented
}

func (s *masterService) CreateService(ctx context.Context, userID int64, input CreateServiceInput) (*domain.Service, error) {
	return nil, errUnimplemented
}

func (s *masterService) UpdateService(ctx context.Context, userID int64, serviceID int64, input UpdateServiceInput) (*domain.Service, error) {
	return nil, errUnimplemented
}

func (s *masterService) ReplaceSchedule(ctx context.Context, userID int64, input ReplaceScheduleInput) ([]domain.SchedulePeriod, error) {
	return nil, errUnimplemented
}

func (s *masterService) GetSchedule(ctx context.Context, userID int64) ([]domain.SchedulePeriod, error) {
	return nil, errUnimplemented
}
