package usecase

import (
	"context"
	"errors"
	"github.com/xmaximix/envilope-chako-server/internal/auth/repository"
	errs "github.com/xmaximix/envilope-chako-server/pkg/error"
	"golang.org/x/crypto/bcrypt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginUser struct {
	repo       *repository.UserRepo
	jwtKey     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewLoginUser(r *repository.UserRepo, jwtKey []byte, accessTTL, refreshTTL time.Duration) *LoginUser {
	return &LoginUser{repo: r, jwtKey: jwtKey, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (uc *LoginUser) Execute(ctx context.Context, email, password string) (*Tokens, error) {
	u, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errs.Wrap("finding user by email", err)
	}
	if !u.Verified {
		return nil, errors.New("email not verified")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errs.Wrap("checking password", err)
	}

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   u.ID.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(uc.accessTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessStr, err := token.SignedString(uc.jwtKey)
	if err != nil {
		return nil, errs.Wrap("signing access token", err)
	}

	refresh := uuid.NewString()
	expires := now.Add(uc.refreshTTL)
	if err := uc.repo.StoreRefreshToken(ctx, u.ID, refresh, expires); err != nil {
		return nil, errs.Wrap("storing refresh token", err)
	}
	return &Tokens{AccessToken: accessStr, RefreshToken: refresh}, nil
}
