package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/aleksandrboro/slotvio/internal/handler/dto"
	"github.com/aleksandrboro/slotvio/internal/service"
)

type AuthHandler struct {
	log         *slog.Logger
	authService AuthService
}

func NewAuthHandler(log *slog.Logger, authService AuthService) *AuthHandler {
	if log == nil {
		log = slog.Default()
	}

	return &AuthHandler{
		log:         log,
		authService: authService,
	}
}

type AuthService interface {
	LoginWithTelegram(ctx context.Context, initData string) (service.AuthData, error)
}

// HandleAuth authenticates Telegram Mini App initData and returns a backend session token.
func (h *AuthHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	log := h.log.With(slog.String("op", "AuthHandler.HandleAuth"))
	initData := r.Header.Get("X-Telegram-Init-Data")
	if initData == "" {
		WriteError(log, w, http.StatusUnauthorized, "unauthorized")
		return
	}

	data, err := h.authService.LoginWithTelegram(r.Context(), initData)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInitData) {
			log.Warn("invalid Telegram init data")
			WriteError(log, w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if errors.Is(err, service.ErrEmptyToken) {
			log.Error("Telegram bot token is not configured")
		}
		log.Error("authentication failed", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	body, err := json.Marshal(dto.AuthResponse{Token: data.Token, ExpAt: data.ExpAt})
	if err != nil {
		log.Error("failed to marshal auth response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		log.Error("failed to write auth response", slog.Any("error", err))
	}
}
