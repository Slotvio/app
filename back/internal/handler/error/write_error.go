package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func WriteError(log *slog.Logger, w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := ErrorMessage{
		Code:    code,
		Message: msg,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to send error message", slog.String("error", err.Error()))
	}
}
