package usecase

import (
	"context"
	"github.com/xmaximix/envilope-chako-server/internal/auth/repository"
	errs "github.com/xmaximix/envilope-chako-server/pkg/error"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type RefreshToken struct {
	repo      *repository.UserRepo
	jwtKey    []byte
	accessTTL time.Duration
}

func NewRefreshToken(r *repository.UserRepo, jwtKey []byte, accessTTL time.Duration) *RefreshToken {
	return &RefreshToken{repo: r, jwtKey: jwtKey, accessTTL: accessTTL}
}

func (uc *RefreshToken) Execute(ctx context.Context, tokenStr string) (string, error) {
	userID, err := uc.repo.ValidateRefreshToken(ctx, tokenStr)
	if err != nil {
		return "", errs.Wrap("validating refresh token", err)
	}
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   uuid.Nil.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(uc.accessTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	claims.Subject = userID.String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.jwtKey)
}
