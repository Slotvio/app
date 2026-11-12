package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aleksandrboro/slotvio/internal/auth"
	"github.com/aleksandrboro/slotvio/internal/config"
	"github.com/aleksandrboro/slotvio/internal/domain"
)

var errAuthRepoTest = errors.New("auth repository test error")

const testBotToken = "123456789:TEST_BOT_TOKEN"

var testTelegramUserID int64 = 123456789

type authUserRepoStub struct {
	findFn   func(context.Context, int64) (*domain.User, error)
	createFn func(context.Context, domain.User) (*domain.User, error)
}

func (r *authUserRepoStub) FindUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	if r.findFn == nil {
		return nil, errors.New("unexpected FindUserByTelegramID call")
	}
	return r.findFn(ctx, telegramID)
}

func (r *authUserRepoStub) CreateUser(ctx context.Context, user domain.User) (*domain.User, error) {
	if r.createFn == nil {
		return nil, errors.New("unexpected CreateUser call")
	}
	return r.createFn(ctx, user)
}

type authSessionRepoStub struct {
	createFn    func(context.Context, domain.AuthSession) (*domain.AuthSession, error)
	findUserFn  func(context.Context, int64) (*domain.AuthSession, error)
	findTokenFn func(context.Context, [32]byte) (int64, error)
}

func (r *authSessionRepoStub) CreateSession(ctx context.Context, session domain.AuthSession) (*domain.AuthSession, error) {
	if r.createFn == nil {
		return nil, errors.New("unexpected CreateSession call")
	}
	return r.createFn(ctx, session)
}

func (r *authSessionRepoStub) FindSessionByUserID(ctx context.Context, userID int64) (*domain.AuthSession, error) {
	if r.findUserFn == nil {
		return nil, errors.New("unexpected FindSessionByUserID call")
	}
	return r.findUserFn(ctx, userID)
}

func (r *authSessionRepoStub) FindSessionByToken(ctx context.Context, tokenHash [32]byte) (int64, error) {
	if r.findTokenFn == nil {
		return 0, errors.New("unexpected FindSessionByToken call")
	}
	return r.findTokenFn(ctx, tokenHash)
}

func signedInitData(t *testing.T, botToken string, telegramID int64) string {
	t.Helper()

	userJSON, err := json.Marshal(map[string]any{
		"id":         telegramID,
		"first_name": "Test",
		"last_name":  "User",
		"username":   "test_user",
	})
	if err != nil {
		t.Fatalf("marshal test Telegram user: %v", err)
	}

	return signedInitDataValues(t, botToken, url.Values{
		"auth_date": {strconv.FormatInt(time.Now().Unix(), 10)},
		"query_id":  {"AAEAAExampleQueryID"},
		"user":      {string(userJSON)},
	})
}

func signedInitDataValues(t *testing.T, botToken string, values url.Values) string {
	t.Helper()
	if values.Get("auth_date") == "" {
		values.Set("auth_date", strconv.FormatInt(time.Now().Unix(), 10))
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, key+"="+values.Get(key))
	}
	dataCheckString := strings.Join(pairs, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secretMAC.Write([]byte(botToken))
	hashMAC := hmac.New(sha256.New, secretMAC.Sum(nil))
	_, _ = hashMAC.Write([]byte(dataCheckString))
	values.Set("hash", hex.EncodeToString(hashMAC.Sum(nil)))

	return values.Encode()
}

func testAuthService(
	token string,
	ttl time.Duration,
	sessions SessionRepo,
	users UserRepo,
) *AuthService {
	return NewAuthService(sessions, users, config.TGCfg{
		Token:      token,
		ExpTime:    time.Minute,
		SessionTTL: ttl,
	})
}

func TestLoginWithTelegramExistingUserCreatesSession(t *testing.T) {
	ctx := context.Background()
	telegramID := testTelegramUserID
	user := &domain.User{ID: 77, TelegramID: telegramID}
	var createdSession domain.AuthSession
	createdUser := false

	users := &authUserRepoStub{
		findFn: func(_ context.Context, gotTelegramID int64) (*domain.User, error) {
			if gotTelegramID != telegramID {
				t.Fatalf("FindUserByTelegramID() id = %d, want %d", gotTelegramID, telegramID)
			}
			return user, nil
		},
		createFn: func(context.Context, domain.User) (*domain.User, error) {
			createdUser = true
			return nil, nil
		},
	}
	sessions := &authSessionRepoStub{
		createFn: func(_ context.Context, session domain.AuthSession) (*domain.AuthSession, error) {
			createdSession = session
			return &session, nil
		},
	}

	result, err := testAuthService(testBotToken, 24*time.Hour, sessions, users).LoginWithTelegram(
		ctx,
		signedInitData(t, testBotToken, telegramID),
	)
	if err != nil {
		t.Fatalf("LoginWithTelegram() error = %v", err)
	}
	if createdUser {
		t.Fatal("CreateUser() called for an existing user")
	}
	if result.UserID != user.ID {
		t.Errorf("UserID = %d, want %d", result.UserID, user.ID)
	}
	if result.Token == "" {
		t.Fatal("LoginWithTelegram() returned an empty token")
	}
	if auth.HashSessionToken(result.Token) != createdSession.TokenHash {
		t.Error("stored session hash does not match returned token")
	}
	if createdSession.UserID != user.ID {
		t.Errorf("created session UserID = %d, want %d", createdSession.UserID, user.ID)
	}
	if !createdSession.ExpiresAt.Equal(result.ExpAt) {
		t.Errorf("session expiry = %v, response expiry = %v", createdSession.ExpiresAt, result.ExpAt)
	}
	if !result.ExpAt.After(time.Now()) || result.ExpAt.After(time.Now().Add(24*time.Hour+time.Second)) {
		t.Errorf("unexpected session expiry: %v", result.ExpAt)
	}
	if createdSession.CreatedAt.IsZero() {
		t.Error("session CreatedAt was not set")
	}
}

func TestLoginWithTelegramCreatesMissingUser(t *testing.T) {
	telegramID := testTelegramUserID
	var createdUser domain.User
	users := &authUserRepoStub{
		findFn: func(_ context.Context, gotTelegramID int64) (*domain.User, error) {
			if gotTelegramID != telegramID {
				t.Fatalf("FindUserByTelegramID() id = %d, want %d", gotTelegramID, telegramID)
			}
			return nil, nil
		},
		createFn: func(_ context.Context, user domain.User) (*domain.User, error) {
			createdUser = user
			user.ID = 88
			return &user, nil
		},
	}
	sessions := &authSessionRepoStub{
		createFn: func(_ context.Context, session domain.AuthSession) (*domain.AuthSession, error) {
			if session.UserID != 88 {
				t.Errorf("session UserID = %d, want 88", session.UserID)
			}
			return &session, nil
		},
	}

	result, err := testAuthService(testBotToken, time.Hour, sessions, users).LoginWithTelegram(
		context.Background(), signedInitData(t, testBotToken, telegramID),
	)
	if err != nil {
		t.Fatalf("LoginWithTelegram() error = %v", err)
	}
	if createdUser.TelegramID != telegramID || createdUser.FirstName != "Test" || createdUser.LastName != "User" || createdUser.Username != "test_user" {
		t.Errorf("created user fields = %#v", createdUser)
	}
	if createdUser.CreatedAt.IsZero() || !createdUser.CreatedAt.Equal(createdUser.UpdatedAt) {
		t.Errorf("created user timestamps not set consistently: %#v", createdUser)
	}
	if result.UserID != 88 {
		t.Errorf("result UserID = %d, want 88", result.UserID)
	}
}

func TestLoginWithTelegramRejectsInvalidInitDataBeforeRepositoryCalls(t *testing.T) {
	users := &authUserRepoStub{}
	sessions := &authSessionRepoStub{}
	invalidData := signedInitData(t, testBotToken, testTelegramUserID) + "tampered"

	_, err := testAuthService(testBotToken, time.Hour, sessions, users).LoginWithTelegram(context.Background(), invalidData)
	if !errors.Is(err, ErrInvalidInitData) {
		t.Errorf("LoginWithTelegram() error = %v, want ErrInvalidInitData", err)
	}
}

func TestLoginWithTelegramReturnsParseErrorForSignedMalformedUser(t *testing.T) {
	users := &authUserRepoStub{}
	sessions := &authSessionRepoStub{}
	data := signedInitDataValues(t, testBotToken, url.Values{
		"user": {"not-json"},
	})

	_, err := testAuthService(testBotToken, time.Minute, sessions, users).LoginWithTelegram(context.Background(), data)
	if !errors.Is(err, ErrFailedToParse) {
		t.Errorf("LoginWithTelegram() error = %v, want ErrFailedToParse", err)
	}
}

func TestLoginWithTelegramRejectsEmptyBotToken(t *testing.T) {
	users := &authUserRepoStub{}
	sessions := &authSessionRepoStub{}

	_, err := testAuthService("", time.Hour, sessions, users).LoginWithTelegram(context.Background(), "anything")
	if !errors.Is(err, ErrEmptyToken) {
		t.Errorf("LoginWithTelegram() error = %v, want ErrEmptyToken", err)
	}
}

func TestLoginWithTelegramRejectsNonPositiveSessionTTL(t *testing.T) {
	user := &domain.User{ID: 77, TelegramID: testTelegramUserID}
	users := &authUserRepoStub{
		findFn: func(context.Context, int64) (*domain.User, error) { return user, nil },
	}
	sessions := &authSessionRepoStub{}

	_, err := testAuthService(testBotToken, 0, sessions, users).LoginWithTelegram(
		context.Background(), signedInitData(t, testBotToken, testTelegramUserID),
	)
	if err == nil || !strings.Contains(err.Error(), "session TTL") {
		t.Errorf("LoginWithTelegram() error = %v, want invalid session TTL", err)
	}
}

func TestLoginWithTelegramWrapsUserRepositoryError(t *testing.T) {
	users := &authUserRepoStub{
		findFn: func(context.Context, int64) (*domain.User, error) { return nil, errAuthRepoTest },
	}
	sessions := &authSessionRepoStub{}

	_, err := testAuthService(testBotToken, time.Hour, sessions, users).LoginWithTelegram(
		context.Background(), signedInitData(t, testBotToken, testTelegramUserID),
	)
	if !errors.Is(err, errAuthRepoTest) {
		t.Errorf("LoginWithTelegram() error = %v, want wrapped repository error", err)
	}
}

func TestLoginWithTelegramWrapsCreateUserError(t *testing.T) {
	users := &authUserRepoStub{
		findFn: func(context.Context, int64) (*domain.User, error) { return nil, nil },
		createFn: func(context.Context, domain.User) (*domain.User, error) {
			return nil, errAuthRepoTest
		},
	}
	sessions := &authSessionRepoStub{}

	_, err := testAuthService(testBotToken, time.Hour, sessions, users).LoginWithTelegram(
		context.Background(), signedInitData(t, testBotToken, testTelegramUserID),
	)
	if !errors.Is(err, errAuthRepoTest) {
		t.Errorf("LoginWithTelegram() error = %v, want wrapped repository error", err)
	}
}

func TestLoginWithTelegramWrapsCreateSessionError(t *testing.T) {
	user := &domain.User{ID: 77, TelegramID: testTelegramUserID}
	users := &authUserRepoStub{
		findFn: func(context.Context, int64) (*domain.User, error) { return user, nil },
	}
	sessions := &authSessionRepoStub{
		createFn: func(context.Context, domain.AuthSession) (*domain.AuthSession, error) {
			return nil, errAuthRepoTest
		},
	}

	_, err := testAuthService(testBotToken, time.Hour, sessions, users).LoginWithTelegram(
		context.Background(), signedInitData(t, testBotToken, testTelegramUserID),
	)
	if !errors.Is(err, errAuthRepoTest) {
		t.Errorf("LoginWithTelegram() error = %v, want wrapped repository error", err)
	}
}

func TestAuthenticateSessionRejectsEmptyToken(t *testing.T) {
	sessions := &authSessionRepoStub{}
	service := testAuthService(testBotToken, time.Hour, sessions, &authUserRepoStub{})

	_, err := service.AuthenticateSession(context.Background(), "")
	if !errors.Is(err, ErrEmptyToken) {
		t.Errorf("AuthenticateSession() error = %v, want ErrEmptyToken", err)
	}
}

func TestAuthenticateSessionLooksUpHashedToken(t *testing.T) {
	const token = "opaque-session-token"
	const wantUserID int64 = 77
	expectedHash := auth.HashSessionToken(token)
	called := false
	sessions := &authSessionRepoStub{
		findTokenFn: func(_ context.Context, hash [32]byte) (int64, error) {
			called = true
			if hash != expectedHash {
				t.Errorf("FindSessionByToken() hash = %x, want %x", hash, expectedHash)
			}
			return wantUserID, nil
		},
	}
	service := testAuthService(testBotToken, time.Hour, sessions, &authUserRepoStub{})

	got, err := service.AuthenticateSession(context.Background(), token)
	if err != nil {
		t.Fatalf("AuthenticateSession() error = %v", err)
	}
	if !called {
		t.Fatal("FindSessionByToken() was not called")
	}
	if got != wantUserID {
		t.Errorf("AuthenticateSession() = %d, want %d", got, wantUserID)
	}
}

func TestAuthenticateSessionWrapsRepositoryError(t *testing.T) {
	sessions := &authSessionRepoStub{
		findTokenFn: func(context.Context, [32]byte) (int64, error) {
			return 0, errAuthRepoTest
		},
	}
	service := testAuthService(testBotToken, time.Hour, sessions, &authUserRepoStub{})

	_, err := service.AuthenticateSession(context.Background(), "opaque-session-token")
	if !errors.Is(err, errAuthRepoTest) {
		t.Errorf("AuthenticateSession() error = %v, want wrapped repository error", err)
	}
}
