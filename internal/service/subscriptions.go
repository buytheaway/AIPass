package service

import (
	"context"
	"time"

	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/repository"
	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo *repository.Store
}

func (s *SubscriptionService) Assign(ctx context.Context, userID, planID uuid.UUID, startsAt *time.Time, status domain.SubscriptionStatus) (domain.UserSubscription, error) {
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		return domain.UserSubscription{}, mapRepoError(err)
	}
	if !plan.IsActive {
		return domain.UserSubscription{}, ErrInvalidInput
	}
	start := time.Now().UTC()
	if startsAt != nil {
		start = startsAt.UTC()
	}
	if status == "" {
		status = domain.SubscriptionPendingPayment
	}
	now := time.Now().UTC()
	sub := domain.UserSubscription{
		ID:        uuid.New(),
		UserID:    userID,
		PlanID:    planID,
		StartsAt:  start,
		EndsAt:    start.AddDate(0, 0, plan.DurationDays),
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.CreateSubscription(ctx, sub)
}

func (s *SubscriptionService) Get(ctx context.Context, id uuid.UUID) (domain.UserSubscription, error) {
	sub, err := s.repo.GetSubscription(ctx, id)
	return sub, mapRepoError(err)
}

func (s *SubscriptionService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserSubscription, error) {
	return s.repo.ListSubscriptionsByUser(ctx, userID)
}

func (s *SubscriptionService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SubscriptionStatus) (domain.UserSubscription, error) {
	sub, err := s.repo.UpdateSubscriptionStatus(ctx, id, status)
	return sub, mapRepoError(err)
}
