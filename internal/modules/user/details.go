package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonyarianto/gobete/internal/systems/db"
	"github.com/sonyarianto/gobete/internal/systems/response"
	"gorm.io/gorm"
)

func GetCurrentUserHandler(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64)) // JWT stores numbers as float64

	// Fetch user details from the database
	var user User
	if err := db.DB.Select("id, email").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.SendErrorResponse(c, fiber.StatusNotFound, "not_found")
		}
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "db_error", "Failed to query user")
	}

	var userDetail UserDetail
	if err := db.DB.Select("id, first_name, last_name, user_id").Where("user_id = ?", user.ID).First(&userDetail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.SendErrorResponse(c, fiber.StatusNotFound, "not_found")
		}
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "db_error", "Failed to query user details")
	}

	// Return user details
	return response.SendSuccessResponse(c, "User details fetched successfully", fiber.Map{
		"id":         user.ID,
		"first_name": userDetail.FirstName,
		"last_name":  userDetail.LastName,
		"email":      user.Email,
	})
}
