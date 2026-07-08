package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/aipass/aipass/internal/domain"
	"github.com/aipass/aipass/internal/repository"
	"github.com/google/uuid"
)

type QRService struct {
	repo *repository.Store
}

type GeneratedQRPass struct {
	Pass  domain.QRPass
	Token string
}

func (s *QRService) Generate(ctx context.Context, subscriptionID uuid.UUID) (GeneratedQRPass, error) {
	sub, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return GeneratedQRPass{}, mapRepoError(err)
	}
	token, err := randomToken()
	if err != nil {
		return GeneratedQRPass{}, err
	}
	now := time.Now().UTC()
	pass := domain.QRPass{
		ID:             uuid.New(),
		UserID:         sub.UserID,
		SubscriptionID: sub.ID,
		TokenHash:      HashQRToken(token),
		Status:         domain.QRPassActive,
		ExpiresAt:      sub.EndsAt,
		CreatedAt:      now,
	}
	created, err := s.repo.CreateQRPass(ctx, pass)
	if err != nil {
		return GeneratedQRPass{}, err
	}
	return GeneratedQRPass{Pass: created, Token: token}, nil
}

func (s *QRService) GetLatestByUser(ctx context.Context, userID uuid.UUID) (domain.QRPass, error) {
	pass, err := s.repo.GetLatestQRPassByUser(ctx, userID)
	return pass, mapRepoError(err)
}

func (s *QRService) Revoke(ctx context.Context, id uuid.UUID) error {
	return mapRepoError(s.repo.RevokeQRPass(ctx, id))
}

func HashQRToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
