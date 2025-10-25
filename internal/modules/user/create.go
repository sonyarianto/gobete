package user

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sonyarianto/gobete/internal/systems/db"
	"github.com/sonyarianto/gobete/internal/systems/response"
	"github.com/sonyarianto/gobete/internal/systems/utility"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateUserHandler(c *fiber.Ctx) error {
	var req CreateUserRequest

	req = *c.Locals("body").(*CreateUserRequest) // Get parsed body from context, after middleware parsing

	// Validate input
	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := utility.FormatValidationErrors(req, validationErrors)
			return response.SendErrorResponse(c, fiber.StatusBadRequest, "validation_error", errors)
		}
		return response.SendErrorResponse(c, fiber.StatusBadRequest, "validation_error", err.Error())
	}

	// Check if user already exists
	var existing User
	if err := db.DB.Select("id").Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return response.SendErrorResponse(c, fiber.StatusBadRequest, "user_exists")
	}

	// Hash password, use bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "Failed to hash password")
	}
	req.Password = string(hashedPassword)

	user := User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	detail := UserDetail{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Transaction to create user and user detail
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// Create user
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// Create user detail
		detail.UserID = user.ID
		if err := tx.Create(&detail).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "db_error", "Failed to create user")
	}

	// Return success response with user ID
	return response.SendSuccessResponse(c, "User created successfully", fiber.Map{
		"id": user.ID,
	})
}
