package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/aleksandrboro/slotvio/internal/auth"
	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/aleksandrboro/slotvio/internal/handler/dto"
	requestid "github.com/aleksandrboro/slotvio/internal/request_id"
	"github.com/aleksandrboro/slotvio/internal/service"
	"github.com/gorilla/mux"
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
	) (*service.PublicProfileOutput, error)

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

	ReplaceSchedule(
		ctx context.Context,
		userID int64,
		input service.ReplaceScheduleInput,
	) ([]domain.SchedulePeriod, error)

	GetSchedule(
		ctx context.Context,
		userID int64,
	) ([]domain.SchedulePeriod, error)
}

// POST /api/v1/master/profile; Create master profile handler
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

	master, err := h.masterService.CreateProfile(ctx, userID, input)
	if err != nil {
		log.Error("failed to create profile", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}
	if master == nil {
		WriteError(log, w, http.StatusNotFound, "masters profile not found")
		return
	}

	resp := dto.CreateProfileResponse{
		ID:                master.ID,
		Slug:              master.Slug,
		Title:             master.Title,
		Category:          master.Category,
		Description:       master.Description,
		Timezone:          master.Timezone,
		SubscriptionUntil: master.SubscriptionUntil,
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

// GET /api/v1/master/profile; Get masters own profile
func (h *MasterHandler) HandleOwnProfile(w http.ResponseWriter, r *http.Request) {
	op := "MasterHandler.HandleOwnProfile"
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

	master, err := h.masterService.GetOwnProfile(ctx, userID)
	if err != nil {
		log.Error("failed to get masters own profile", slog.String("error", err.Error()))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}
	if master == nil {
		WriteError(log, w, http.StatusNotFound, "masters profile not found")
		return
	}

	resp := dto.OwnProfileResponse{
		Slug:        master.Slug,
		Title:       master.Title,
		Category:    master.Category,
		Description: master.Description,
		Timezone:    master.Timezone,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal get own profile response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("profile received", slog.String("master", resp.Slug))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.String("master", resp.Slug),
			slog.Any("error", err),
		)
	}
}

// GET /api/v1/masters/{slug}; Get masters public profile
func (h *MasterHandler) HandlePublicProfile(w http.ResponseWriter, r *http.Request) {
	op := "MasterHandler.HandlePublicProfile"
	log := h.log.With(slog.String("op", op))
	ctx := r.Context()

	requestID, err := requestid.RequestIDFromContext(ctx)
	if err != nil {
		log.Error("failed to get request_id from context", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}
	log = log.With(slog.String("request_id", string(requestID)))

	slug, ok := mux.Vars(r)["slug"]
	if !ok {
		WriteError(log, w, http.StatusBadRequest, "provide masters slug")
		return
	}

	profileData, err := h.masterService.GetPublicProfile(ctx, slug)
	if err != nil {
		log.Error("failed to get masters public profile", slog.String("error", err.Error()))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}
	if profileData == nil {
		WriteError(log, w, http.StatusNotFound, "masters profile not found")
		return
	}

	resp := dto.PublicProfileResponse{
		Slug:        profileData.Master.Slug,
		Title:       profileData.Master.Title,
		Category:    profileData.Master.Category,
		Description: profileData.Master.Description,
		Timezone:    profileData.Master.Timezone,
		Services:    make([]dto.ServiceInfo, 0, len(profileData.Services)),
	}

	for _, s := range profileData.Services {
		resp.Services = append(resp.Services, dto.ServiceInfo{
			ID:              s.ID,
			Name:            s.Name,
			Description:     s.Description,
			DurationMinutes: s.DurationMinutes,
			Price:           s.Price,
		})
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal get own profile response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("profile received", slog.String("master", resp.Slug))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.String("master", resp.Slug),
			slog.Any("error", err),
		)
	}
}

// PATCH /api/v1/master/profile; Update masters profile
func (h *MasterHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	op := "MasterHandler.HandleUpdateProfile"
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

	var req dto.UpdateProfileRequest
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

	noProfileFields := req.Title == nil && req.Category == nil && req.Description == nil && req.Timezone == nil
	if noProfileFields {
		WriteError(log, w, http.StatusBadRequest, "at least one profile field must be provided")
		return
	}

	input := service.UpdateProfileInput{
		Title:       req.Title,
		Category:    req.Category,
		Description: req.Description,
		Timezone:    req.Timezone,
	}

	master, err := h.masterService.UpdateProfile(ctx, userID, input)
	if err != nil {
		log.Error("failed to update masters profile", slog.String("error", err.Error()))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}
	if master == nil {
		WriteError(log, w, http.StatusNotFound, "masters profile not found")
		return
	}

	resp := dto.UpdateProfileResponse{
		ID:          master.ID,
		Slug:        master.Slug,
		Title:       master.Title,
		Category:    master.Category,
		Description: master.Description,
		Timezone:    master.Timezone,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal update profile response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("profile updated", slog.String("master", resp.Slug))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.String("master", resp.Slug),
			slog.Any("error", err),
		)
	}
}

// GET /api/v1/master/services; Get masters services
func (h *MasterHandler) HandleListServices(w http.ResponseWriter, r *http.Request) {
	op := "MasterHandler.HandleListServices"
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

	services, err := h.masterService.ListServices(ctx, userID)
	if err != nil {
		log.Error("failed to get masters services", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := dto.ListServiceResponse{
		Services: make([]dto.ServiceInfo, 0, len(services)),
	}

	for _, s := range services {
		resp.Services = append(resp.Services, dto.ServiceInfo{
			ID:              s.ID,
			Name:            s.Name,
			Description:     s.Description,
			DurationMinutes: s.DurationMinutes,
			Price:           s.Price,
		})
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal list services response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("services received", slog.Int64("user_id", userID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
	}
}

// POST /api/v1/master/services; Create new service
func (h *MasterHandler) HandleCreateService(w http.ResponseWriter, r *http.Request) {
	op := "MasterHandler.HandleCreateService"
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

	var req dto.CreateServiceRequest
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

	input := service.CreateServiceInput{
		Name:            req.Name,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
	}

	service, err := h.masterService.CreateService(ctx, userID, input)
	if err != nil {
		log.Error("failed to create service", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}
	if service == nil {
		log.Error("create service returned a nil result without an error")
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := dto.CreateServiceResponse{
		ID:              service.ID,
		Name:            service.Name,
		Description:     service.Description,
		DurationMinutes: service.DurationMinutes,
		Price:           service.Price,
		IsActive:        service.IsActive,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal create service response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("service created", slog.Int64("user_id", userID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
	}
}

// PATCH /api/v1/master/services/{id}; Change or deactivate service
func (h *MasterHandler) HandleUpdateService(w http.ResponseWriter, r *http.Request) {
	op := "MasterHandler.HandleUpdateService"
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

	id, ok := mux.Vars(r)["id"]
	if !ok {
		WriteError(log, w, http.StatusBadRequest, "provide masters slug")
		return
	}

	serviceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Error("failed to parse service id", slog.String("error", err.Error()))
		WriteError(log, w, http.StatusUnauthorized, "internal error")
		return
	}

	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	var req dto.UpdateServiceRequest
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

	noServiceFields := req.Name == nil && req.Description == nil && req.DurationMinutes == nil && req.Price == nil && req.IsActive == nil
	if noServiceFields {
		WriteError(log, w, http.StatusBadRequest, "at least one service field must be provided")
		return
	}

	input := service.UpdateServiceInput{
		Name:            req.Name,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
		IsActive:        req.IsActive,
	}

	service, err := h.masterService.UpdateService(ctx, userID, serviceID, input)
	if err != nil {
		log.Error("failed to update service", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}
	if service == nil {
		log.Error("update service returned a nil result without an error")
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := dto.UpdateServiceResponse{
		ID:              service.ID,
		Name:            service.Name,
		Description:     service.Description,
		DurationMinutes: service.DurationMinutes,
		Price:           service.Price,
		IsActive:        service.IsActive,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal update service response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("service updated", slog.Int64("user_id", userID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
	}
}

// PUT /api/v1/master/schedule; Create or replace schedule
func (h *MasterHandler) HandleReplaceSchedule(w http.ResponseWriter, r *http.Request) {
	op := "MasterHandler.HandleReplaceSchedule"
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

	var req dto.ReplaceScheduleRequest
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

	input := service.ReplaceScheduleInput{
		Periods: make([]service.SchedulePeriodInput, 0, len(req.Days)),
	}

	for _, sp := range req.Days {
		input.Periods = append(input.Periods, service.SchedulePeriodInput{
			DayOfWeek: sp.DayOfWeek,
			StartTime: sp.StartTime,
			EndTime:   sp.EndTime,
		})
	}

	schedule, err := h.masterService.ReplaceSchedule(ctx, userID, input)
	if err != nil {
		log.Error("failed to replace schedule", slog.String("error", err.Error()))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := dto.ScheduleResponse{
		Days: make([]dto.SchedulePeriod, 0, len(schedule)),
	}

	for _, sp := range schedule {
		resp.Days = append(resp.Days, dto.SchedulePeriod{
			DayOfWeek: sp.DayOfWeek,
			StartTime: sp.StartTime.Format("15:04"),
			EndTime:   sp.EndTime.Format("15:04"),
		})
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal replace schedule response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("schedule replaced", slog.Int64("user_id", userID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
	}
}

// GET /api/v1/master/schedule; Get masters schedule
func (h *MasterHandler) HandleGetSchedule(w http.ResponseWriter, r *http.Request) {
	op := "MasterHandler.HandleGetSchedule"
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

	schedule, err := h.masterService.GetSchedule(ctx, userID)
	if err != nil {
		log.Error("failed to get schedule", slog.String("error", err.Error()))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := dto.ScheduleResponse{
		Days: make([]dto.SchedulePeriod, 0, len(schedule)),
	}

	for _, sp := range schedule {
		resp.Days = append(resp.Days, dto.SchedulePeriod{
			DayOfWeek: sp.DayOfWeek,
			StartTime: sp.StartTime.Format("15:04"),
			EndTime:   sp.EndTime.Format("15:04"),
		})
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal get schedule response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("schedule received", slog.Int64("user_id", userID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.Int64("user_id", userID),
			slog.Any("error", err),
		)
	}
}
