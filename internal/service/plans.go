package service

import (
	"context"
	"time"

	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PlanService struct {
	repo *repository.Store
}

func (s *PlanService) Create(ctx context.Context, name string, description *string, durationDays int, price decimal.Decimal, currency string) (domain.SubscriptionPlan, error) {
	if currency == "" {
		currency = "KZT"
	}
	now := time.Now().UTC()
	plan := domain.SubscriptionPlan{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		DurationDays: durationDays,
		Price:        price,
		Currency:     currency,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	return s.repo.CreatePlan(ctx, plan)
}

func (s *PlanService) Get(ctx context.Context, id uuid.UUID) (domain.SubscriptionPlan, error) {
	plan, err := s.repo.GetPlan(ctx, id)
	return plan, mapRepoError(err)
}

func (s *PlanService) List(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	return s.repo.ListPlans(ctx)
}

func (s *PlanService) Update(ctx context.Context, id uuid.UUID, name *string, description *string, durationDays *int, price *decimal.Decimal, currency *string, isActive *bool) (domain.SubscriptionPlan, error) {
	plan, err := s.repo.UpdatePlan(ctx, id, name, description, durationDays, price, currency, isActive)
	return plan, mapRepoError(err)
}

func (s *PlanService) Deactivate(ctx context.Context, id uuid.UUID) (domain.SubscriptionPlan, error) {
	active := false
	plan, err := s.repo.UpdatePlan(ctx, id, nil, nil, nil, nil, nil, &active)
	return plan, mapRepoError(err)
}
