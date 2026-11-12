package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/aleksandrboro/slotvio/internal/config"
	"github.com/aleksandrboro/slotvio/internal/handler/dto"
	"github.com/aleksandrboro/slotvio/internal/service"
)

type AuthHandler struct {
	log         *slog.Logger
	authService AuthService
}

func NewAuthHandler(log *slog.Logger, authService AuthService, tgCfg config.TGCfg) *AuthHandler {
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

func (h *AuthHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	op := "handler.HandleAuth"
	log := h.log.With(slog.String("op", op))

	ctx := r.Context()

	initData := r.Header.Get("X-Telegram-Init-Data")
	if initData == "" {
		WriteError(log, w, http.StatusUnauthorized, "unauthorized")
		return
	}

	data, err := h.authService.LoginWithTelegram(ctx, initData)
	if err != nil {
		if errors.Is(err, service.ErrEmptyToken) {
			log.Error("empty bot token")
			WriteError(log, w, http.StatusInternalServerError, "internal error")
			return
		}
		if errors.Is(err, service.ErrInvalidInitData) {
			log.Warn("invalid Telegram init data", slog.String("error", err.Error()))
			WriteError(log, w, http.StatusUnauthorized, "unauthorized")
			return
		}
		log.Error("authentication failed", slog.String("error", err.Error()))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := dto.AuthResponse{
		Token: data.Token,
		ExpAt: data.ExpAt,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error("failed to marshal auth response", slog.Any("error", err))
		WriteError(log, w, http.StatusInternalServerError, "internal error")
		return
	}

	log.Info("user authenticated", slog.Int64("user_id", data.UserID))

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(body); err != nil {
		log.Error("failed to send response",
			slog.Int64("user_id", data.UserID),
			slog.Any("error", err),
		)
	}
}
