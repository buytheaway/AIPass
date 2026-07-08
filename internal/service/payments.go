package service

import (
	"context"
	"time"

	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentService struct {
	repo *repository.Store
}

func (s *PaymentService) CreateManual(ctx context.Context, subscriptionID uuid.UUID, amount decimal.Decimal, method domain.PaymentMethod, receiptFileID *uuid.UUID) (domain.Payment, error) {
	sub, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return domain.Payment{}, mapRepoError(err)
	}
	now := time.Now().UTC()
	payment := domain.Payment{
		ID:             uuid.New(),
		UserID:         sub.UserID,
		SubscriptionID: subscriptionID,
		Amount:         amount,
		Currency:       "KZT",
		Method:         method,
		Status:         domain.PaymentUploaded,
		ReceiptFileID:  receiptFileID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return s.repo.CreatePayment(ctx, payment)
}

func (s *PaymentService) List(ctx context.Context) ([]domain.Payment, error) {
	return s.repo.ListPayments(ctx)
}

func (s *PaymentService) Get(ctx context.Context, id uuid.UUID) (domain.Payment, error) {
	payment, err := s.repo.GetPayment(ctx, id)
	return payment, mapRepoError(err)
}

func (s *PaymentService) Approve(ctx context.Context, id uuid.UUID, adminID uuid.UUID) (domain.Payment, error) {
	payment, err := s.repo.UpdatePaymentStatus(ctx, id, domain.PaymentApproved, &adminID)
	return payment, mapRepoError(err)
}

func (s *PaymentService) Reject(ctx context.Context, id uuid.UUID) (domain.Payment, error) {
	payment, err := s.repo.UpdatePaymentStatus(ctx, id, domain.PaymentRejected, nil)
	return payment, mapRepoError(err)
}
