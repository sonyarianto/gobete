package home

import (
	"github.com/sonyarianto/gobete/internal/systems/response"

	"os"

	"github.com/gofiber/fiber/v2"
)

func HomeHandler(c *fiber.Ctx) error {
	return response.SendSuccessResponse(c, "API is running.", fiber.Map{
		"version": os.Getenv("APP_VERSION"),
	})
}
