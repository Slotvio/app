package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aleksandrboro/slotvio/internal/service"
)

var errAuthHandlerTest = errors.New("auth service test error")

type authServiceStub struct {
	loginFn func(context.Context, string) (service.AuthData, error)
}

func (s *authServiceStub) LoginWithTelegram(ctx context.Context, initData string) (service.AuthData, error) {
	if s.loginFn == nil {
		return service.AuthData{}, errors.New("unexpected LoginWithTelegram call")
	}
	return s.loginFn(ctx, initData)
}

func newAuthHandlerForTest(svc AuthService) *AuthHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewAuthHandler(logger, svc)
}

func TestHandleAuthSuccess(t *testing.T) {
	wantExpiry := time.Date(2027, time.January, 2, 3, 4, 5, 0, time.UTC)
	var gotInitData string
	stub := &authServiceStub{
		loginFn: func(_ context.Context, initData string) (service.AuthData, error) {
			gotInitData = initData
			return service.AuthData{UserID: 42, Token: "opaque-session-token", ExpAt: wantExpiry}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", nil)
	req.Header.Set("X-Telegram-Init-Data", "signed-telegram-init-data")
	rec := httptest.NewRecorder()
	newAuthHandlerForTest(stub).HandleAuth(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if gotInitData != "signed-telegram-init-data" {
		t.Errorf("initData = %q, want header value", gotInitData)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var got struct {
		Token   string    `json:"token"`
		Expires time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Token != "opaque-session-token" {
		t.Errorf("token = %q, want opaque-session-token", got.Token)
	}
	if !got.Expires.Equal(wantExpiry) {
		t.Errorf("expires_at = %s, want %s", got.Expires, wantExpiry)
	}
}

func TestHandleAuthRejectsMissingInitData(t *testing.T) {
	called := false
	stub := &authServiceStub{
		loginFn: func(context.Context, string) (service.AuthData, error) {
			called = true
			return service.AuthData{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", nil)
	rec := httptest.NewRecorder()
	newAuthHandlerForTest(stub).HandleAuth(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if called {
		t.Fatal("LoginWithTelegram() called without initData")
	}
}

func TestHandleAuthMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{name: "invalid Telegram initData", serviceErr: service.ErrInvalidInitData, wantStatus: http.StatusUnauthorized},
		{name: "unexpected service failure", serviceErr: errAuthHandlerTest, wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &authServiceStub{
				loginFn: func(context.Context, string) (service.AuthData, error) {
					return service.AuthData{}, tt.serviceErr
				},
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram", nil)
			req.Header.Set("X-Telegram-Init-Data", "provided-init-data")
			rec := httptest.NewRecorder()

			newAuthHandlerForTest(stub).HandleAuth(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
