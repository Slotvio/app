package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aleksandrboro/slotvio/internal/auth"
	"github.com/aleksandrboro/slotvio/internal/config"
	"github.com/aleksandrboro/slotvio/internal/domain"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

var (
	ErrInvalidInitData = errors.New("invalid init data")
	ErrEmptyToken      = errors.New("telegram bot token is empty")
	ErrFailedToParse   = errors.New("failed to parse init data")
)

type AuthService struct {
	sessionRepo SessionRepo
	userRepo    UserRepo
	tgCfg       config.TGCfg
}

func NewAuthService(sessionRepo SessionRepo, userRepo UserRepo, tgCfg config.TGCfg) *AuthService {
	return &AuthService{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		tgCfg:       tgCfg,
	}
}

type AuthData struct {
	UserID int64
	Token  string
	ExpAt  time.Time
}

type SessionRepo interface {
	CreateSession(ctx context.Context, session domain.AuthSession) (*domain.AuthSession, error)
	FindSessionByUserID(ctx context.Context, userID int64) (*domain.AuthSession, error)
	FindSessionByToken(ctx context.Context, hashToken [32]byte) (userID int64, err error)
}

type UserRepo interface {
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)
	FindUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
}

func (s *AuthService) LoginWithTelegram(
	ctx context.Context,
	initData string,
) (AuthData, error) {
	if s.tgCfg.Token == "" {
		return AuthData{}, ErrEmptyToken
	}

	if err := initdata.Validate(initData, s.tgCfg.Token, s.tgCfg.ExpTime); err != nil {
		return AuthData{}, ErrInvalidInitData
	}

	data, err := initdata.Parse(initData)
	if err != nil {
		return AuthData{}, ErrFailedToParse
	}

	user, err := s.userRepo.FindUserByTelegramID(ctx, data.User.ID)
	if err != nil {
		return AuthData{}, fmt.Errorf("find user for authentication: %w", err)
	}

	if user == nil {
		now := time.Now().UTC()
		user, err = s.userRepo.CreateUser(ctx, domain.User{
			TelegramID: data.User.ID,
			FirstName:  data.User.FirstName,
			LastName:   data.User.LastName,
			Username:   data.User.Username,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
		if err != nil {
			return AuthData{}, fmt.Errorf("create user for authentication: %w", err)
		}
		if user == nil {
			return AuthData{}, errors.New("user repository returned nil user")
		}
	}

	if s.tgCfg.SessionTTL <= 0 {
		return AuthData{}, errors.New("session TTL must be positive")
	}

	token, tokenHash, err := auth.GenerateSessionToken()
	if err != nil {
		return AuthData{}, fmt.Errorf("generate session token: %w", err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.tgCfg.SessionTTL)

	_, err = s.sessionRepo.CreateSession(ctx, domain.AuthSession{
		UserID:    user.ID,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return AuthData{}, fmt.Errorf("create authentication session: %w", err)
	}

	return AuthData{
		UserID: user.ID,
		Token:  token,
		ExpAt:  expiresAt,
	}, nil
}

func (s *AuthService) AuthenticateSession(ctx context.Context, token string) (int64, error) {
	if token == "" {
		return 0, ErrEmptyToken
	}

	hashToken := auth.HashSessionToken(token)

	userID, err := s.sessionRepo.FindSessionByToken(ctx, hashToken)
	if err != nil {
		return 0, fmt.Errorf("find session for authentication: %w", err)
	}

	return userID, nil
}
