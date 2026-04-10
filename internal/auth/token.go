package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService issues and validates JWT tokens.
type TokenService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewTokenService creates a TokenService with the given HMAC secret and TTLs.
func NewTokenService(secret string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// IssueAccessToken creates a short-lived JWT for API access.
func (s *TokenService) IssueAccessToken(userID string) (string, error) {
	return s.issueToken(userID, s.accessTTL)
}

// IssueRefreshToken creates a longer-lived JWT for refreshing access tokens.
func (s *TokenService) IssueRefreshToken(userID string) (string, error) {
	return s.issueToken(userID, s.refreshTTL)
}

func (s *TokenService) issueToken(userID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ValidateToken parses and validates a JWT, returning the user ID from the subject claim.
func (s *TokenService) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.Subject, nil
}
