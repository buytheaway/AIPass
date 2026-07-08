package service

import (
	"database/sql"
	"errors"

	"github.com/aipass/aipass/internal/auth"
	"github.com/aipass/aipass/internal/config"
	"github.com/aipass/aipass/internal/repository"
	"go.uber.org/zap"
)

type Container struct {
	Auth          *AuthService
	Users         *UserService
	Plans         *PlanService
	Subscriptions *SubscriptionService
	QR            *QRService
	Access        *AccessService
	Payments      *PaymentService
	Reports       *ReportService
}

func NewContainer(cfg config.Config, repos *repository.Store, log *zap.Logger) *Container {
	tokenManager, err := auth.NewTokenManager(
		cfg.Auth.PrivateKeyPEM,
		cfg.Auth.PublicKeyPEM,
		cfg.Auth.PrivateKeyPath,
		cfg.Auth.PublicKeyPath,
		cfg.Auth.AccessTokenTTL,
	)
	if err != nil {
		log.Warn("jwt keys are not configured; auth login will return auth_not_configured", zap.Error(err))
	}

	users := &UserService{repo: repos}
	plans := &PlanService{repo: repos}
	subs := &SubscriptionService{repo: repos}
	qr := &QRService{repo: repos}
	access := &AccessService{repo: repos, log: log}
	payments := &PaymentService{repo: repos}
	reports := &ReportService{repo: repos}

	return &Container{
		Auth:          &AuthService{repo: repos, tokens: tokenManager},
		Users:         users,
		Plans:         plans,
		Subscriptions: subs,
		QR:            qr,
		Access:        access,
		Payments:      payments,
		Reports:       reports,
	}
}

func mapRepoError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
