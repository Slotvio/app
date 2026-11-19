package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/aleksandrboro/slotvio/internal/auth"
	"github.com/aleksandrboro/slotvio/internal/domain"
	"github.com/aleksandrboro/slotvio/internal/handler/dto"
	"github.com/aleksandrboro/slotvio/internal/service"
	"github.com/gorilla/mux"
)

type BookingHandler struct {
	log            *slog.Logger
	bookingService BookingService
}

func NewBookingHandler(log *slog.Logger, bookingService BookingService) *BookingHandler {
	if log == nil {
		log = slog.Default()
	}
	return &BookingHandler{log: log, bookingService: bookingService}
}

type BookingService interface {
	GetAvailableSlots(ctx context.Context, slug string, serviceID int64, date string) (service.AvailableSlots, error)
	CreateBooking(ctx context.Context, clientID int64, input service.CreateBookingInput) (*domain.Booking, error)
}

// HandleListSlots handles GET /api/v1/masters/{slug}/slots?date=YYYY-MM-DD&service_id=id.
func (h *BookingHandler) HandleListSlots(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("op", "BookingHandler.HandleListSlots"))
	slug, ok := mux.Vars(r)["slug"]
	if !ok {
		WriteError(log, w, http.StatusBadRequest, "provide master slug")
		return
	}

	serviceID, err := strconv.ParseInt(r.URL.Query().Get("service_id"), 10, 64)
	if err != nil || serviceID <= 0 {
		WriteError(log, w, http.StatusBadRequest, "service_id must be a positive integer")
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		WriteError(log, w, http.StatusBadRequest, "provide date in YYYY-MM-DD format")
		return
	}

	available, err := h.bookingService.GetAvailableSlots(r.Context(), slug, serviceID, date)
	if err != nil {
		h.writeServiceError(log, w, err, false)
		return
	}

	response := dto.AvailableSlotsResponse{
		Date:     available.Date,
		Timezone: available.Timezone,
		Slots:    make([]dto.AvailableSlot, 0, len(available.Slots)),
	}
	for _, slot := range available.Slots {
		response.Slots = append(response.Slots, dto.AvailableSlot{
			StartAt:   slot.StartAt,
			LocalTime: slot.LocalTime,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

// HandleCreateBooking handles POST /api/v1/bookings. The client ID is taken
// from the authenticated session, never from the request body.
func (h *BookingHandler) HandleCreateBooking(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("op", "BookingHandler.HandleCreateBooking"))
	clientID, ok := auth.UserIDFromContext(r.Context())
	if !ok || clientID <= 0 {
		WriteError(log, w, http.StatusUnauthorized, "unauthorized")
		return
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)

	var request dto.CreateBookingRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
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

	booking, err := h.bookingService.CreateBooking(r.Context(), clientID, service.CreateBookingInput{
		MasterSlug:    request.MasterSlug,
		ServiceID:     request.ServiceID,
		StartAt:       request.StartAt,
		ClientComment: request.ClientComment,
	})
	if err != nil {
		h.writeServiceError(log, w, err, true)
		return
	}
	if booking == nil {
		log.Error("booking service returned nil booking without error")
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, dto.CreateBookingResponse{
		ID:              booking.ID,
		MasterID:        booking.MasterID,
		ServiceID:       booking.ServiceID,
		ServiceName:     booking.ServiceNameSnapshot,
		DurationMinutes: booking.DurationMinutesSnapshot,
		Price:           booking.PriceSnapshot,
		StartAt:         booking.StartAt,
		EndAt:           booking.EndAt,
		Status:          booking.Status,
	})
}

func (h *BookingHandler) writeServiceError(log *slog.Logger, w http.ResponseWriter, err error, conflictIsExpected bool) {
	switch {
	case errors.Is(err, service.ErrInvalidBookingInput):
		WriteError(log, w, http.StatusBadRequest, "invalid booking request")
	case errors.Is(err, service.ErrBookableResourceNotFound):
		WriteError(log, w, http.StatusNotFound, "master or active service not found")
	case conflictIsExpected && errors.Is(err, service.ErrBookingSlotUnavailable):
		WriteError(log, w, http.StatusConflict, "booking slot is no longer available")
	default:
		log.Error("booking operation failed", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
