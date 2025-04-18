package usecase

import (
	"context"
	"github.com/xmaximix/envilope-chako-server/internal/auth/domain"
	"github.com/xmaximix/envilope-chako-server/internal/auth/repository"
	"github.com/xmaximix/envilope-chako-server/pkg/email"
	errs "github.com/xmaximix/envilope-chako-server/pkg/error"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUser struct {
	repo   *repository.UserRepo
	sender email.Sender
}

func NewRegisterUser(r *repository.UserRepo, s email.Sender) *RegisterUser {
	return &RegisterUser{repo: r, sender: s}
}

func (uc *RegisterUser) Execute(ctx context.Context, emailAddr, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errs.Wrap("hashing password", err)
	}

	u := &domain.User{
		Email:        emailAddr,
		PasswordHash: string(hash),
		Verified:     false,
		Role:         domain.RoleUser,
	}
	if err := uc.repo.Create(ctx, u); err != nil {
		return errs.Wrap("creating user", err)
	}

	code := uuid.NewString()
	expires := time.Now().Add(24 * time.Hour)
	if err := uc.repo.StoreRefreshToken(ctx, u.ID, code, expires); err != nil {
		return errs.Wrap("storing refresh token", err)
	}

	subject := "Your verification code"
	return uc.sender.Send(ctx, emailAddr, subject, code)
}
