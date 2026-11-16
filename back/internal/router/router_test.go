package router

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aleksandrboro/slotvio/internal/handler"
	"github.com/aleksandrboro/slotvio/internal/service"
)

type routerAuthStub struct {
	initData string
	login    service.AuthData
	err      error
}

func (s *routerAuthStub) LoginWithTelegram(_ context.Context, initData string) (service.AuthData, error) {
	s.initData = initData
	return s.login, s.err
}

func (s *routerAuthStub) AuthenticateSession(context.Context, string) (int64, error) {
	return 1, nil
}

func newRouterAuthStub() *routerAuthStub {
	return &routerAuthStub{
		login: service.AuthData{
			UserID: 1,
			Token:  "test-session-token",
			ExpAt:  time.Date(2027, time.January, 2, 3, 4, 5, 0, time.UTC),
		},
	}
}

func TestNewRouterRejectsNilRequiredHandlers(t *testing.T) {
	stub := newRouterAuthStub()
	authHandler := handler.NewAuthHandler(nil, stub)
	masterHandler := handler.NewMasterHandler(nil, nil)

	tests := []struct {
		name          string
		masterHandler *handler.MasterHandler
		authHandler   *handler.AuthHandler
		wantErr       error
	}{
		{
			name:        "nil master handler",
			authHandler: authHandler,
			wantErr:     errEmptyMasterHandler,
		},
		{
			name:          "nil auth handler",
			masterHandler: masterHandler,
			wantErr:       errEmptyAuthHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRouter(tt.masterHandler, tt.authHandler, stub)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewRouter() error = %v, want %v", err, tt.wantErr)
			}
			if got != nil {
				t.Fatal("NewRouter() returned a router on error")
			}
		})
	}
}

func TestNewRouterRegistersTelegramAuthRoute(t *testing.T) {
	stub := newRouterAuthStub()
	r, err := NewRouter(
		handler.NewMasterHandler(nil, nil),
		handler.NewAuthHandler(nil, stub),
		stub,
	)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", nil)
	req.Header.Set("X-Telegram-Init-Data", "valid-init-data")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if stub.initData != "valid-init-data" {
		t.Errorf("LoginWithTelegram() initData = %q, want valid-init-data", stub.initData)
	}

	var got struct {
		Token   string    `json:"token"`
		Expires time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Token != "test-session-token" {
		t.Errorf("token = %q, want test-session-token", got.Token)
	}
	if !got.Expires.Equal(stub.login.ExpAt) {
		t.Errorf("expires_at = %s, want %s", got.Expires, stub.login.ExpAt)
	}
}

func TestNewRouterProtectsMasterRoutes(t *testing.T) {
	stub := newRouterAuthStub()
	r, err := NewRouter(
		handler.NewMasterHandler(nil, nil),
		handler.NewAuthHandler(nil, stub),
		stub,
	)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/master/profile", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}
