package config

import (
	"errors"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:      false,
		ErrorHandler: CustomError(),
		JSONEncoder:  sonic.ConfigStd.Marshal,
		JSONDecoder:  sonic.ConfigStd.Unmarshal,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	return app
}

func CustomError() fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		code := http.StatusInternalServerError
		if err, ok := err.(*fiber.Error); ok {
			code = err.Code
		}

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
		}

		message := &Message{
			Success: false,
			Message: err.Error(),
		}

		return ctx.Status(code).JSON(message)
	}
}

type Message struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
