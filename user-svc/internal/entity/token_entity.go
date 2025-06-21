package entity

import (
	"time"
)

type AdminAccessToken struct {
	UserID    string
	Token     string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	ExpiresAt time.Time
}

type UserAccessToken struct {
	UserID    string
	Token     string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	ExpiresAt time.Time
}
