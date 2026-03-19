package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleClient = "client"
	RoleMaster = "master"
	RoleAdmin  = "admin"
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