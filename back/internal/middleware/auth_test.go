package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aleksandrboro/slotvio/internal/auth"
)

var errSessionAuthTest = errors.New("session auth test error")

type sessionAuthenticatorFunc func(context.Context, string) (int64, error)

func (f sessionAuthenticatorFunc) AuthenticateSession(ctx context.Context, token string) (int64, error) {
	return f(ctx, token)
}

func TestAuthMiddlewareRejectsMissingOrMalformedBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
	}{
		{name: "missing header"},
		{name: "wrong scheme", authorization: "Basic abc"},
		{name: "missing token", authorization: "Bearer"},
		{name: "extra fields", authorization: "Bearer abc extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCalls := 0
			authenticator := sessionAuthenticatorFunc(func(context.Context, string) (int64, error) {
				authCalls++
				return 7, nil
			})
			nextCalls := 0
			handler := AuthMiddleware(authenticator)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				nextCalls++
			}))

			req := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
			if authCalls != 0 {
				t.Errorf("AuthenticateSession() calls = %d, want 0", authCalls)
			}
			if nextCalls != 0 {
				t.Errorf("next handler calls = %d, want 0", nextCalls)
			}
		})
	}
}

func TestAuthMiddlewareRejectsInvalidSession(t *testing.T) {
	nextCalls := 0
	authenticator := sessionAuthenticatorFunc(func(_ context.Context, token string) (int64, error) {
		if token != "expired-token" {
			t.Errorf("token = %q, want expired-token", token)
		}
		return 0, errSessionAuthTest
	})
	handler := AuthMiddleware(authenticator)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalls++
	}))

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if nextCalls != 0 {
		t.Errorf("next handler calls = %d, want 0", nextCalls)
	}
}

func TestAuthMiddlewareAddsAuthenticatedUserIDToContext(t *testing.T) {
	const wantUserID int64 = 73
	const wantToken = "valid-session-token"
	authenticator := sessionAuthenticatorFunc(func(_ context.Context, token string) (int64, error) {
		if token != wantToken {
			t.Errorf("token = %q, want %q", token, wantToken)
		}
		return wantUserID, nil
	})
	nextCalls := 0
	handler := AuthMiddleware(authenticator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls++
		gotUserID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			t.Error("user ID missing from request context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if gotUserID != wantUserID {
			t.Errorf("context user ID = %d, want %d", gotUserID, wantUserID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer "+wantToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if nextCalls != 1 {
		t.Errorf("next handler calls = %d, want 1", nextCalls)
	}
}
