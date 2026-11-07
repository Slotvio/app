package handler

import "log/slog"

type MasterHandler struct {
	log           *slog.Logger
	masterService MasterService
}

type MasterService interface {
	CreateProfile() (string, error)
}
