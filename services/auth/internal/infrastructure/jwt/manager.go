package jwt

import (
	"context"
	"time"

	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/application/ports"
	"github.com/BladeRunner322/orange-team-microservices/services/auth/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewManager(secret, issuer, audience string, ttl time.Duration) *Manager {
	return &Manager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (m *Manager) Generate(ctx context.Context, userID string, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  []string{m.audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Validate(ctx context.Context, tokenString string) (ports.UserInfo, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return ports.UserInfo{}, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return ports.UserInfo{}, domain.ErrInvalidToken
	}
	return ports.UserInfo{
		UserID: claims.UserID,
		Role:   claims.Role,
	}, nil
}
