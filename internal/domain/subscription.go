package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SubscriptionStatus string

const (
	SubscriptionPendingPayment SubscriptionStatus = "pending_payment"
	SubscriptionActive         SubscriptionStatus = "active"
	SubscriptionExpired        SubscriptionStatus = "expired"
	SubscriptionCancelled      SubscriptionStatus = "cancelled"
)

type SubscriptionPlan struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	Name         string          `db:"name" json:"name"`
	Description  *string         `db:"description" json:"description,omitempty"`
	DurationDays int             `db:"duration_days" json:"duration_days"`
	Price        decimal.Decimal `db:"price" json:"price"`
	Currency     string          `db:"currency" json:"currency"`
	IsActive     bool            `db:"is_active" json:"is_active"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at" json:"updated_at"`
}

type UserSubscription struct {
	ID        uuid.UUID          `db:"id" json:"id"`
	UserID    uuid.UUID          `db:"user_id" json:"user_id"`
	PlanID    uuid.UUID          `db:"plan_id" json:"plan_id"`
	StartsAt  time.Time          `db:"starts_at" json:"starts_at"`
	EndsAt    time.Time          `db:"ends_at" json:"ends_at"`
	Status    SubscriptionStatus `db:"status" json:"status"`
	CreatedAt time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt time.Time          `db:"updated_at" json:"updated_at"`
}
