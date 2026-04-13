package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleClient UserRole = "client"
	RoleMaster UserRole = "master"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         string
	AvatarURL    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
