package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aipass/aipass/internal/auth"
	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/repository"
	"github.com/google/uuid"
)

type UserService struct {
	repo *repository.Store
}

type CreateUserInput struct {
	Email    string
	Phone    *string
	FullName string
	Role     domain.Role
	Password *string
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (domain.User, error) {
	now := time.Now().UTC()
	var passwordHash *string
	if input.Password != nil {
		hash, err := auth.HashPassword(*input.Password)
		if err != nil {
			return domain.User{}, err
		}
		passwordHash = &hash
	}
	user := domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		Phone:        input.Phone,
		FullName:     input.FullName,
		Role:         input.Role,
		PasswordHash: passwordHash,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	return user, mapRepoError(err)
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, phone *string, fullName *string, isActive *bool) (domain.User, error) {
	user, err := s.repo.UpdateUser(ctx, id, phone, fullName, isActive)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, ErrNotFound
	}
	return user, err
}
