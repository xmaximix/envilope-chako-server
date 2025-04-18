package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Verified     bool      `db:"verified"`
	Role         Role      `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
}
