package auth

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/aipass/aipass/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Email  string      `json:"email"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	ttl        time.Duration
}

func NewTokenManager(privateKeyPEM, publicKeyPEM, privateKeyPath, publicKeyPath string, ttl time.Duration) (*TokenManager, error) {
	if privateKeyPEM != "" || publicKeyPEM != "" {
		if privateKeyPEM == "" || publicKeyPEM == "" {
			return nil, errors.New("both JWT_PRIVATE_KEY_PEM and JWT_PUBLIC_KEY_PEM must be set")
		}
		return NewTokenManagerFromPEM([]byte(privateKeyPEM), []byte(publicKeyPEM), ttl)
	}

	privateBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	publicBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	return NewTokenManagerFromPEM(privateBytes, publicBytes, ttl)
}

func NewTokenManagerFromPEM(privateBytes, publicBytes []byte, ttl time.Duration) (*TokenManager, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	if err != nil {
		return nil, err
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	if err != nil {
		return nil, err
	}
	return &TokenManager{privateKey: privateKey, publicKey: publicKey, ttl: ttl}, nil
}

func (m *TokenManager) Generate(user domain.User) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(m.privateKey)
}

func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return m.publicKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
