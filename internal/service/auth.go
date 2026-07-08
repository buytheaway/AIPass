package service

import (
	"context"

	"github.com/aipass/aipass/internal/auth"
	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/repository"
)

type AuthService struct {
	repo   *repository.Store
	tokens *auth.TokenManager
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, domain.User, error) {
	if s.tokens == nil {
		return "", domain.User{}, ErrAuthNotConfigured
	}
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", domain.User{}, ErrUnauthorized
	}
	if user.PasswordHash == nil || !auth.CheckPassword(password, *user.PasswordHash) {
		return "", domain.User{}, ErrUnauthorized
	}
	if !user.IsActive {
		return "", domain.User{}, ErrForbidden
	}
	token, err := s.tokens.Generate(user)
	return token, user, err
}

func (s *AuthService) TokenManager() *auth.TokenManager {
	return s.tokens
}
