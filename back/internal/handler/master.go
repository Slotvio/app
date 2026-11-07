package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/aleksandrboro/slotvio/internal/auth"
	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/aleksandrboro/slotvio/internal/handler/dto"
	requestid "github.com/aleksandrboro/slotvio/internal/request_id"
	"github.com/aleksandrboro/slotvio/internal/service"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9_-]{3,40}$`)

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
	op := "MasterHandler.HandleCreateProfile"
	log := h.log.With(slog.String("op", op))
	ctx := r.Context()

	requestID, err := requestid.RequestIDFromContext(ctx)
	if err != nil {
		log.Error("failed to get request_id from context", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}
	log = log.With(slog.String("request_id", string(requestID)))

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		WriteError(log, w, http.StatusUnauthorized, "unauthorized")
		return
	}

	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	var req dto.CreateProfileRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			WriteError(log, w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		WriteError(log, w, http.StatusBadRequest, "failed to parse request body")
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		WriteError(log, w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !slugPattern.MatchString(req.Slug) {
		WriteError(log, w, http.StatusBadRequest, "invalid slug")
		return
	}

	if _, err := time.LoadLocation(req.Timezone); err != nil {
		WriteError(log, w, http.StatusBadRequest, "invalid timezone")
		return
	}

	input := service.CreateProfileInput{
		Slug:        req.Slug,
		Title:       req.Title,
		Category:    req.Category,
		Description: req.Description,
		Timezone:    req.Timezone,
	}

	resp, err := h.masterService.CreateProfile(ctx, userID, input)
	if err != nil {
		log.Error("failed to create profile", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal create profile response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("profile created", slog.String("master", resp.Slug))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.String("master", resp.Slug),
			slog.Any("error", err),
		)
	}
}
