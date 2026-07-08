package dto

import (
	"time"

	"github.com/aipass/aipass/internal/domain"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	User        domain.User `json:"user"`
}

type CreateUserRequest struct {
	Email    string      `json:"email" validate:"required,email"`
	Phone    *string     `json:"phone"`
	FullName string      `json:"full_name" validate:"required,min=2"`
	Role     domain.Role `json:"role" validate:"required,oneof=admin member"`
	Password *string     `json:"password" validate:"omitempty,min=8"`
}

type UpdateUserRequest struct {
	Phone    *string `json:"phone"`
	FullName *string `json:"full_name" validate:"omitempty,min=2"`
	IsActive *bool   `json:"is_active"`
}

type CreatePlanRequest struct {
	Name         string  `json:"name" validate:"required,min=2"`
	Description  *string `json:"description"`
	DurationDays int     `json:"duration_days" validate:"required,min=1"`
	Price        string  `json:"price" validate:"required"`
	Currency     string  `json:"currency"`
}

type UpdatePlanRequest struct {
	Name         *string `json:"name" validate:"omitempty,min=2"`
	Description  *string `json:"description"`
	DurationDays *int    `json:"duration_days" validate:"omitempty,min=1"`
	Price        *string `json:"price"`
	Currency     *string `json:"currency"`
	IsActive     *bool   `json:"is_active"`
}

type AssignSubscriptionRequest struct {
	PlanID   uuid.UUID                 `json:"plan_id" validate:"required"`
	StartsAt *time.Time                `json:"starts_at"`
	Status   domain.SubscriptionStatus `json:"status" validate:"omitempty,oneof=pending_payment active expired cancelled"`
}

type UpdateSubscriptionStatusRequest struct {
	Status domain.SubscriptionStatus `json:"status" validate:"required,oneof=pending_payment active expired cancelled"`
}

type QRPassResponse struct {
	Pass  domain.QRPass `json:"pass"`
	Token string        `json:"token,omitempty"`
}

type ValidateScanRequest struct {
	QRToken   string `json:"qr_token" validate:"required"`
	ScannerID string `json:"scanner_id" validate:"required"`
}

type ValidateScanResponse struct {
	Decision  domain.AccessDecision  `json:"decision"`
	EventType domain.AccessEventType `json:"event_type"`
	User      *ScanUser              `json:"user,omitempty"`
	Reason    *string                `json:"reason"`
}

type ScanUser struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
}
