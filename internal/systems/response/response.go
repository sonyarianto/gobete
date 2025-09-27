package response

import (
	"github.com/gofiber/fiber/v2"
	errpkg "github.com/sonyarianto/gobete/internal/systems/error"
)

// Success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
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
	var msg string
	var detail any

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
		default:
			// Use error message from map for the code
			if def, ok := errpkg.ErrorMessages[code]; ok {
				msg = def
			} else {
				msg = "Unknown error"
			}
			detail = m
		}
	} else if def, ok := errpkg.ErrorMessages[code]; ok {
		msg = def
	} else {
		msg = "Unknown error"
	}

	resp := ErrorResponse{
		Code:    code,
		Success: false,
		Message: msg,
		Details: detail,
	}

	return c.Status(status).JSON(resp)
}
