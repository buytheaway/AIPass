package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/repository"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type EventPublisher interface {
	Publish(ctx context.Context, key string, value []byte) error
}

type AccessService struct {
	repo   *repository.Store
	log    *zap.Logger
	Redis  *goredis.Client
	Events EventPublisher
}

type ScanResult struct {
	Decision  domain.AccessDecision
	EventType domain.AccessEventType
	User      *domain.User
	Reason    *string
}

type AccessEventMessage struct {
	EventID        uuid.UUID              `json:"event_id"`
	EventType      domain.AccessEventType `json:"event_type"`
	Decision       domain.AccessDecision  `json:"decision"`
	Reason         *string                `json:"reason"`
	UserID         uuid.UUID              `json:"user_id"`
	SubscriptionID *uuid.UUID             `json:"subscription_id"`
	QRPassID       *uuid.UUID             `json:"qr_pass_id"`
	ScannerID      string                 `json:"scanner_id"`
	OccurredAt     time.Time              `json:"occurred_at"`
}

func (s *AccessService) ValidateScan(ctx context.Context, rawToken string, scannerID string) (ScanResult, error) {
	if scannerID == "" || rawToken == "" {
		return ScanResult{}, ErrInvalidInput
	}
	if allowed, err := s.checkRateLimit(ctx, scannerID); err == nil && !allowed {
		reason := "rate_limited"
		return ScanResult{Decision: domain.DecisionDenied, EventType: domain.AccessDenied, Reason: &reason}, nil
	}

	tokenHash := HashQRToken(rawToken)
	record, err := s.repo.GetQRValidationRecord(ctx, tokenHash)
	if err != nil {
		reason := "qr_not_found"
		return ScanResult{Decision: domain.DecisionDenied, EventType: domain.AccessDenied, Reason: &reason}, nil
	}

	now := time.Now().UTC()
	if reason := denyReason(record, now); reason != nil {
		subID := record.Subscription.ID
		passID := record.Pass.ID
		event, createErr := s.createEvent(ctx, &record.User.ID, &subID, &passID, domain.AccessDenied, domain.DecisionDenied, reason, scannerID, now)
		if createErr == nil {
			s.publish(ctx, event)
		}
		return ScanResult{Decision: domain.DecisionDenied, EventType: domain.AccessDenied, Reason: reason}, nil
	}

	latest, err := s.repo.LatestAllowedAccessEvent(ctx, record.User.ID)
	if err != nil {
		return ScanResult{}, err
	}
	eventType := domain.AccessCheckIn
	if latest != nil && latest.EventType == domain.AccessCheckIn {
		eventType = domain.AccessCheckOut
	}

	subID := record.Subscription.ID
	passID := record.Pass.ID
	event, err := s.createEvent(ctx, &record.User.ID, &subID, &passID, eventType, domain.DecisionAllowed, nil, scannerID, now)
	if err != nil {
		return ScanResult{}, err
	}
	s.publish(ctx, event)

	return ScanResult{
		Decision:  domain.DecisionAllowed,
		EventType: eventType,
		User:      &record.User,
	}, nil
}

func denyReason(record repository.QRValidationRecord, now time.Time) *string {
	var reason string
	switch {
	case record.Pass.Status != domain.QRPassActive:
		reason = "qr_not_active"
	case now.After(record.Pass.ExpiresAt):
		reason = "qr_expired"
	case !record.User.IsActive:
		reason = "user_inactive"
	case record.Subscription.Status != domain.SubscriptionActive:
		reason = "subscription_not_active"
	case now.Before(record.Subscription.StartsAt) || now.After(record.Subscription.EndsAt):
		reason = "subscription_out_of_period"
	default:
		return nil
	}
	return &reason
}

func (s *AccessService) createEvent(ctx context.Context, userID *uuid.UUID, subID *uuid.UUID, passID *uuid.UUID, eventType domain.AccessEventType, decision domain.AccessDecision, reason *string, scannerID string, now time.Time) (domain.AccessEvent, error) {
	if userID == nil {
		empty := uuid.Nil
		userID = &empty
	}
	event := domain.AccessEvent{
		ID:             uuid.New(),
		UserID:         *userID,
		SubscriptionID: subID,
		QRPassID:       passID,
		EventType:      eventType,
		Decision:       decision,
		Reason:         reason,
		ScannerID:      &scannerID,
		OccurredAt:     now,
		CreatedAt:      now,
	}
	return s.repo.CreateAccessEvent(ctx, event)
}

func (s *AccessService) publish(ctx context.Context, event domain.AccessEvent) {
	if s.Events == nil {
		return
	}
	scannerID := ""
	if event.ScannerID != nil {
		scannerID = *event.ScannerID
	}
	payload, err := json.Marshal(AccessEventMessage{
		EventID: event.ID, EventType: event.EventType, Decision: event.Decision, Reason: event.Reason,
		UserID: event.UserID, SubscriptionID: event.SubscriptionID, QRPassID: event.QRPassID,
		ScannerID: scannerID, OccurredAt: event.OccurredAt,
	})
	if err != nil {
		s.log.Error("marshal access event", zap.Error(err))
		return
	}
	if err := s.Events.Publish(ctx, event.UserID.String(), payload); err != nil {
		s.log.Error("publish access event", zap.Error(err))
	}
}

func (s *AccessService) checkRateLimit(ctx context.Context, scannerID string) (bool, error) {
	if s.Redis == nil {
		return true, nil
	}
	key := "rate_limit:scanner:" + scannerID
	count, err := s.Redis.Incr(ctx, key).Result()
	if err != nil {
		return true, err
	}
	if count == 1 {
		_ = s.Redis.Expire(ctx, key, time.Minute).Err()
	}
	return count <= 120, nil
}
