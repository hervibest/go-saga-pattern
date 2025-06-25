package model

import (
	"net"

	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	Username  string    `json:"username" validate:"required,min=0,max=255"`
	Email     string    `json:"email" validate:"required,min=0,max=255,email"`
	Password  string    `json:"password" validate:"required,min=6"`
	CreatedBy uuid.UUID `json:"created_by" validate:"required"`
	UpdatedBy uuid.UUID `json:"updated_by" validate:"required"`
	RequestID uuid.UUID `json:"request_id" validate:"required"`
	IpAddress net.IP    `json:"ip_address" validate:"required"`
}

func (r *RegisterUserRequest) SetRequestIDAndIpAddress(requestID uuid.UUID, ipAddress net.IP) {
	r.RequestID = requestID
	r.IpAddress = ipAddress
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,min=0,max=255,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
