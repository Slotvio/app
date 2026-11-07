package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/aleksandrboro/slotvio/internal/service"
)

type MasterHandler struct {
	log           *slog.Logger
	masterService MasterService
}

func NewMasterHandler(log *slog.Logger, masterService MasterService) *MasterHandler {
	if log == nil {
		log = slog.Default()
	}

	return &MasterHandler{
		log:           log,
		masterService: masterService,
	}
}

type MasterService interface {
	CreateProfile(
		ctx context.Context,
		userID int64,
		input service.CreateProfileInput,
	) (*domain.Master, error)

	GetOwnProfile(
		ctx context.Context,
		userID int64,
	) (*domain.Master, error)

	GetPublicProfile(
		ctx context.Context,
		slug string,
	) (*domain.Master, error)

	UpdateProfile(
		ctx context.Context,
		userID int64,
		input service.UpdateProfileInput,
	) (*domain.Master, error)

	ListServices(
		ctx context.Context,
		userID int64,
	) ([]domain.Service, error)

	CreateService(
		ctx context.Context,
		userID int64,
		input service.CreateServiceInput,
	) (*domain.Service, error)

	UpdateService(
		ctx context.Context,
		userID int64,
		serviceID int64,
		input service.UpdateServiceInput,
	) (*domain.Service, error)

	GetSchedule(
		ctx context.Context,
		userID int64,
	) ([]domain.SchedulePeriod, error)

	ReplaceSchedule(
		ctx context.Context,
		userID int64,
		input service.ReplaceScheduleInput,
	) ([]domain.SchedulePeriod, error)
}

func (h *MasterHandler) HandleCreateProfile(w http.ResponseWriter, r *http.Request) {
	// op := "MasterHandler.HandleCreateProfile"

	// log := h.log.With(slog.String("op", op))
}
