package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type User struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	Phone        *string    `db:"phone" json:"phone,omitempty"`
	FullName     string     `db:"full_name" json:"full_name"`
	Role         Role       `db:"role" json:"role"`
	PasswordHash *string    `db:"password_hash" json:"-"`
	PhotoFileID  *uuid.UUID `db:"photo_file_id" json:"photo_file_id,omitempty"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
