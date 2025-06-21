package model

import (
	"net"

	"github.com/google/uuid"
)

type RegisterUserRequest struct {
	Username  string    `json:"username" validate:"required"`
	Email     string    `json:"email" validate:"required"`
	Password  string    `json:"password" validate:"required"`
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
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
