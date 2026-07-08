package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccessEventType string
type AccessDecision string

const (
	AccessCheckIn  AccessEventType = "check_in"
	AccessCheckOut AccessEventType = "check_out"
	AccessDenied   AccessEventType = "denied"

	DecisionAllowed AccessDecision = "allowed"
	DecisionDenied  AccessDecision = "denied"
)

type AccessEvent struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	UserID         uuid.UUID       `db:"user_id" json:"user_id"`
	SubscriptionID *uuid.UUID      `db:"subscription_id" json:"subscription_id,omitempty"`
	QRPassID       *uuid.UUID      `db:"qr_pass_id" json:"qr_pass_id,omitempty"`
	EventType      AccessEventType `db:"event_type" json:"event_type"`
	Decision       AccessDecision  `db:"decision" json:"decision"`
	Reason         *string         `db:"reason" json:"reason,omitempty"`
	ScannerID      *string         `db:"scanner_id" json:"scanner_id,omitempty"`
	PhotoFileID    *uuid.UUID      `db:"photo_file_id" json:"photo_file_id,omitempty"`
	OccurredAt     time.Time       `db:"occurred_at" json:"occurred_at"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}
