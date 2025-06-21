package config

import (
	"errors"
	"go-saga-pattern/commoner/utils"
	"log"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:      false,
		AppName:      utils.GetEnv("APP_NAME"),
		ErrorHandler: CustomError(),
		JSONEncoder:  sonic.ConfigStd.Marshal,
		JSONDecoder:  sonic.ConfigStd.Unmarshal,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		stop := time.Now()
		log.Printf("[%s] %s %s - %d (%s) IP: %s",
			stop.Format(time.RFC3339),
			c.Method(),
			c.OriginalURL(),
			c.Response().StatusCode(),
			stop.Sub(start),
			c.IP(),
		)

		return err
	})

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
