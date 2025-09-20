package home

import (
	"github.com/sonyarianto/gobete/internal/systems/response"

	"os"

	"github.com/gofiber/fiber/v2"
)

type Home struct {
	Version string `json:"version" example:"0.0.1"`
}

type HomeSuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"API is running."`
	Data    Home   `json:"data"`
}

// @Summary      Show the home status of API
// @Description  Get the status of the API
// @Tags         home
// @Accept       json
// @Produce      json
// @Success      200  {object}  HomeSuccessResponse
// @Router       / [get]
func HomeHandler(c *fiber.Ctx) error {
	return response.SendSuccessResponse(c, "API is running.", fiber.Map{
		"version": os.Getenv("APP_VERSION"),
	})
}
