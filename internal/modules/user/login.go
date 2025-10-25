package user

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonyarianto/gobete/internal/systems/db"
	"github.com/sonyarianto/gobete/internal/systems/response"
	"github.com/sonyarianto/gobete/internal/systems/utility"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"os"
	"strconv"
	"time"
)

func LoginUserHandler(c *fiber.Ctx) error {
	var req LoginRequest

	req = *c.Locals("body").(*LoginRequest) // Get parsed body from context, after middleware parsing

	// Validate input
	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := utility.FormatValidationErrors(req, validationErrors)
			return response.SendErrorResponse(c, fiber.StatusBadRequest, "validation_error", errors)
		}
		return response.SendErrorResponse(c, fiber.StatusBadRequest, "validation_error", err.Error())
	}

	// Find user by email
	var user User
	if err := db.DB.Select("id, password, email").Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.SendErrorResponse(c, fiber.StatusUnauthorized, "invalid_credentials")
		}
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "db_error", "Failed to query user")
	}

	// Compare password with hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return response.SendErrorResponse(c, fiber.StatusUnauthorized, "invalid_credentials")
	}

	// Will return id, first_name, last_name and email
	var userDetail UserDetail
	if err := db.DB.Select("id, first_name, last_name, user_id").Where("user_id = ?", user.ID).First(&userDetail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.SendErrorResponse(c, fiber.StatusUnauthorized, "record_not_found")
		}
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "db_error", "Failed to query user details")
	}

	// Get env variable for access token expiry
	accessTokenExpireMinutes := os.Getenv("ACCESS_TOKEN_EXPIRE_MINUTES")

	// Convert to int, default to 15 if not set or invalid
	var accessTokenExpire int = 15
	if accessTokenExpireMinutes != "" {
		if v, err := strconv.Atoi(accessTokenExpireMinutes); err == nil {
			accessTokenExpire = v
		}
	}

	// Generate access token (short-lived)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Duration(accessTokenExpire) * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     uuid.NewString(), // Unique identifier for the token
	})

	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "Failed to generate access token")
	}

	// Get env variable for refresh token expiry
	refreshTokenExpireDays := os.Getenv("REFRESH_TOKEN_EXPIRE_DAYS")

	// Convert to int, default to 7 if not set or invalid
	var refreshTokenExpire int = 7
	if refreshTokenExpireDays != "" {
		if v, err := strconv.Atoi(refreshTokenExpireDays); err == nil {
			refreshTokenExpire = v
		}
	}

	// Generate refresh token (long-lived)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Duration(refreshTokenExpire) * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     uuid.NewString(), // Unique identifier for the token
	})

	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "Failed to generate refresh token")
	}

	// Get env variable for session mode
	sessionMode := os.Getenv("SESSION_MODE")

	// If session mode is "jwt_server_stateful", store JTI in UserSession table
	if sessionMode == "jwt_server_stateful" {
		// Get JTI from token claims and store in UserSession
		claims := refreshToken.Claims.(jwt.MapClaims)

		jtiVal, ok := claims["jti"]
		if !ok {
			return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "Failed to get JTI from refresh token claims")
		}
		jti, ok := jtiVal.(string)
		if !ok {
			return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "JTI is not a string")
		}

		issuedAt, ok := claims["iat"].(int64)
		if !ok {
			// fallback for float64 (if token was generated elsewhere)
			if f, ok := claims["iat"].(float64); ok {
				issuedAt = int64(f)
			} else {
				return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "Invalid iat type")
			}
		}

		expiresAt, ok := claims["exp"].(int64)
		if !ok {
			if f, ok := claims["exp"].(float64); ok {
				expiresAt = int64(f)
			} else {
				return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "Invalid exp type")
			}
		}

		// Convert int64 to time.Time
		issuedAtTime := time.Unix(issuedAt, 0)
		expiresAtTime := time.Unix(expiresAt, 0)

		session := UserSession{
			UserID:     user.ID,
			JTI:        jti, // JTI of the refresh token
			CreatedAt:  issuedAtTime,
			ExpiresAt:  expiresAtTime,
			LastSeenAt: issuedAtTime,
		}
		if err := db.DB.Create(&session).Error; err != nil {
			return response.SendErrorResponse(c, fiber.StatusInternalServerError, "db_error", "Failed to create user session")
		}
	}

	// Set refresh token as HttpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshTokenString,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Now().Add(time.Duration(refreshTokenExpire) * 24 * time.Hour),
		Secure:   os.Getenv("ENV") == "production", // Only send cookie over HTTPS in production
		Path:     "/",                              // Cookie valid for all paths
	})

	// Return success response with access token
	return response.SendSuccessResponse(c, "User logged in successfully", fiber.Map{
		"id":           user.ID,
		"first_name":   userDetail.FirstName,
		"last_name":    userDetail.LastName,
		"email":        user.Email,
		"access_token": accessTokenString,
	})
}
