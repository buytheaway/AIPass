package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentMethod string
type PaymentStatus string

const (
	PaymentKaspiManual  PaymentMethod = "kaspi_manual"
	PaymentCash         PaymentMethod = "cash"
	PaymentBankTransfer PaymentMethod = "bank_transfer"

	PaymentUploaded PaymentStatus = "uploaded"
	PaymentApproved PaymentStatus = "approved"
	PaymentRejected PaymentStatus = "rejected"
)

type Payment struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	UserID         uuid.UUID       `db:"user_id" json:"user_id"`
	SubscriptionID uuid.UUID       `db:"subscription_id" json:"subscription_id"`
	Amount         decimal.Decimal `db:"amount" json:"amount"`
	Currency       string          `db:"currency" json:"currency"`
	Method         PaymentMethod   `db:"method" json:"method"`
	Status         PaymentStatus   `db:"status" json:"status"`
	ReceiptFileID  *uuid.UUID      `db:"receipt_file_id" json:"receipt_file_id,omitempty"`
	ApprovedBy     *uuid.UUID      `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt     *time.Time      `db:"approved_at" json:"approved_at,omitempty"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
}
