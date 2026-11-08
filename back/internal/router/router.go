package router

import (
	"errors"
	"net/http"

	"github.com/aleksandrboro/slotvio/internal/handler"
	"github.com/aleksandrboro/slotvio/internal/middleware"
	"github.com/gorilla/mux"
)

var (
	errEmptyMasterHandler = errors.New("master handler is required")
)

func NewRouter(masterHandler *handler.MasterHandler) (*mux.Router, error) {
	if masterHandler == nil {
		return nil, errEmptyMasterHandler
	}

	router := mux.NewRouter()
	router.Use(middleware.RequestID)

	// Публичный маршрут — без auth middleware.
	router.HandleFunc(
		"/api/v1/masters/{slug}",
		masterHandler.HandlePublicProfile,
	).Methods(http.MethodGet)

	// Маршруты личного кабинета мастера — с авторизацией.
	master := router.PathPrefix("/api/v1/master").Subrouter()
	master.Use(middleware.AuthMiddleware)

	master.HandleFunc("/profile", masterHandler.HandleCreateProfile).
		Methods(http.MethodPost)
	master.HandleFunc("/profile", masterHandler.HandleOwnProfile).
		Methods(http.MethodGet)
	master.HandleFunc("/profile", masterHandler.HandleUpdateProfile).
		Methods(http.MethodPatch)

	master.HandleFunc("/services", masterHandler.HandleListServices).
		Methods(http.MethodGet)
	master.HandleFunc("/services", masterHandler.HandleCreateService).
		Methods(http.MethodPost)
	master.HandleFunc("/services/{id}", masterHandler.HandleUpdateService).
		Methods(http.MethodPatch)

	master.HandleFunc("/schedule", masterHandler.HandleGetSchedule).
		Methods(http.MethodGet)
	master.HandleFunc("/schedule", masterHandler.HandleReplaceSchedule).
		Methods(http.MethodPut)

	return router, nil
}
