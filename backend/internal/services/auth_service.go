package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"queryforge/backend/internal/auth"
	"queryforge/backend/internal/config"
	"queryforge/backend/internal/models"
	"queryforge/backend/internal/repository"

	"github.com/google/uuid"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type AuthService struct {
	cfg    config.Config
	users  *repository.UserRepository
	tokens *repository.TokenRepository
}

type TokenPair struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         models.User `json:"user"`
}

func NewAuthService(cfg config.Config, users *repository.UserRepository, tokens *repository.TokenRepository) *AuthService {
	return &AuthService{cfg: cfg, users: users, tokens: tokens}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (TokenPair, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	if name == "" || email == "" || len(password) < 8 {
		return TokenPair{}, errors.New("name, valid email, and password of at least 8 characters are required")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return TokenPair{}, err
	}
	user, err := s.users.Create(ctx, name, email, hash)
	if err != nil {
		return TokenPair{}, err
	}
	return s.issueTokens(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (TokenPair, error) {
	user, err := s.users.FindByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}
	if !auth.CheckPassword(user.PasswordHash, password) {
		return TokenPair{}, ErrInvalidCredentials
	}
	return s.issueTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	claims, err := auth.ParseToken(refreshToken, s.cfg.RefreshSecret, "refresh")
	if err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}
	ok, err := s.tokens.RefreshTokenValid(ctx, claims.UserID, auth.HashToken(refreshToken))
	if err != nil || !ok {
		return TokenPair{}, ErrInvalidCredentials
	}
	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}
	_ = s.tokens.RevokeRefreshToken(ctx, auth.HashToken(refreshToken))
	return s.issueTokens(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.tokens.RevokeRefreshToken(ctx, auth.HashToken(refreshToken))
}

func (s *AuthService) ParseAccessToken(token string) (uuid.UUID, error) {
	claims, err := auth.ParseToken(token, s.cfg.AccessSecret, "access")
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user models.User) (TokenPair, error) {
	access, err := auth.CreateAccessToken(user.ID, user.Email, s.cfg.AccessSecret, s.cfg.AccessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := auth.CreateRefreshToken(user.ID, user.Email, s.cfg.RefreshSecret, s.cfg.RefreshTTL)
	if err != nil {
		return TokenPair{}, err
	}
	expiresAt := time.Now().UTC().Add(s.cfg.RefreshTTL)
	if err := s.tokens.StoreRefreshToken(ctx, user.ID, auth.HashToken(refresh), expiresAt); err != nil {
		return TokenPair{}, err
	}
	user.PasswordHash = ""
	return TokenPair{AccessToken: access, RefreshToken: refresh, User: user}, nil
}
