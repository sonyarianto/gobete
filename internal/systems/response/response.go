package response

import (
	errpkg "github.com/sonyarianto/gobete/internal/systems/error"

	"github.com/gofiber/fiber/v2"
)

// Success response
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Success message."`
	Data    any    `json:"data"`
}

// Error details
type ErrorDetail struct {
	Code    string `json:"code" example:"not_found"`
	Message any    `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Error response
type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Error   ErrorDetail `json:"error"`
}

// Helper for success response
func SendSuccessResponse(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// Helper for error response
func SendErrorResponse(c *fiber.Ctx, status int, code string, message ...any) error {
	var msg any
	var details any

	if len(message) > 0 {
		switch m := message[0].(type) {
		case string:
			if m != "" {
				msg = m
			} else if def, ok := errpkg.ErrorMessages[code]; ok {
				msg = def
			} else {
				msg = "Unknown error"
			}
		case map[string]string:
			msg = errpkg.ErrorMessages[code]
			details = m
		default:
			msg = "Unknown error"
		}
	} else if def, ok := errpkg.ErrorMessages[code]; ok {
		msg = def
	} else {
		msg = "Unknown error"
	}

	resp := fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    code,
			"message": msg,
		},
	}
	// Add type assertion check before setting details
	if details != nil {
		if errMap, ok := resp["error"].(fiber.Map); ok {
			errMap["details"] = details
		}
	}

	return c.Status(status).JSON(resp)
}
