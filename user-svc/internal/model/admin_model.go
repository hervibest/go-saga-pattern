package model

import (
	"net"

	"github.com/google/uuid"
)

type RegisterAdminRequest struct {
	Username  string    `json:"username" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	CreatedBy uuid.UUID `json:"created_by" validate:"required"`
	UpdatedBy uuid.UUID `json:"updated_by" validate:"required"`
	RequestID uuid.UUID `json:"request_id" validate:"required"`
	IpAddress net.IP    `json:"ip_address" validate:"required"`
}

func (r *RegisterAdminRequest) SetRequestIDAndIpAddress(requestID uuid.UUID, ipAddress net.IP) {
	r.RequestID = requestID
	r.IpAddress = ipAddress
}

type LoginAdminRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AdminResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token,omitempty"`
}
