package model

import (
	"time"
)

type AuthResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type LogoutUserRequest struct {
	UserId      string `json:"user_id" validate:"required"`
	AccessToken string
	ExpiresAt   time.Time
}
