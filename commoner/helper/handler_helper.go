package helper

import (
	"errors"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type BaseModel interface {
	SetRequestIDAndIpAddress(requestID uuid.UUID, ipAddress net.IP)
}

func SetBaseModel(ctx *fiber.Ctx, model BaseModel) error {
	parsedIpAddress := net.ParseIP(ctx.IP())
	if parsedIpAddress == nil {
		return errors.New("Invalid IP address")
	}

	requestID := uuid.New()
	model.SetRequestIDAndIpAddress(requestID, parsedIpAddress)
	return nil
}
