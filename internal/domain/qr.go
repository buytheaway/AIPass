package domain

import (
	"time"

	"github.com/google/uuid"
)

type QRPassStatus string

const (
	QRPassActive  QRPassStatus = "active"
	QRPassRevoked QRPassStatus = "revoked"
	QRPassExpired QRPassStatus = "expired"
)

type QRPass struct {
	ID             uuid.UUID    `db:"id" json:"id"`
	UserID         uuid.UUID    `db:"user_id" json:"user_id"`
	SubscriptionID uuid.UUID    `db:"subscription_id" json:"subscription_id"`
	TokenHash      string       `db:"token_hash" json:"-"`
	Status         QRPassStatus `db:"status" json:"status"`
	ExpiresAt      time.Time    `db:"expires_at" json:"expires_at"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
}
