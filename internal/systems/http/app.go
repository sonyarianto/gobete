package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func NewApp() *fiber.App {
	app := fiber.New()

	// Global middlewares
	app.Use(logger.New())

	// Register all routes
	RegisterRoutes(app)

	// 404 handler
	app.Use(NotFoundHandler)

	return app
}
