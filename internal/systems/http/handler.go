package http

import (
	"github.com/sonyarianto/gobete/internal/systems/response"

	"github.com/gofiber/fiber/v2"
)

// 404 Not Found handler
func NotFoundHandler(c *fiber.Ctx) error {
	return response.SendErrorResponse(c, fiber.StatusNotFound, "not_found")
}

func HealthCheckHandler(c *fiber.Ctx) error {
	return response.SendSuccessResponse(c, "API is healthy", fiber.Map{"status": "healthy"})
}
