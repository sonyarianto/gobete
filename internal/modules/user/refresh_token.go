package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonyarianto/gobete/internal/systems/db"
	"github.com/sonyarianto/gobete/internal/systems/response"

	"os"
	"strconv"
	"time"
)

func RefreshTokenHandler(c *fiber.Ctx) error {
	// Get refresh token from cookie
	refreshTokenString := c.Cookies("refresh_token")
	if refreshTokenString == "" {
		return response.SendErrorResponse(c, fiber.StatusUnauthorized, "no_refresh_token")
	}

	// Parse and validate the refresh token
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected_signing_method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return response.SendErrorResponse(c, fiber.StatusUnauthorized, "invalid_refresh_token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return response.SendErrorResponse(c, fiber.StatusUnauthorized, "invalid_refresh_token_claims")
	}

	// Extract user_id and jti
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return response.SendErrorResponse(c, fiber.StatusUnauthorized, "invalid_user_id")
	}
	userID := uint(userIDFloat)

	jti, ok := claims["jti"].(string)
	if !ok {
		return response.SendErrorResponse(c, fiber.StatusUnauthorized, "invalid_jti")
	}

	// Check session mode
	sessionMode := os.Getenv("SESSION_MODE")
	if sessionMode == "jwt_server_stateful" {
		// Check if session exists and is valid
		var session UserSession
		err := db.DB.Where("user_id = ? AND jti = ? AND expires_at > ?", userID, jti, time.Now()).First(&session).Error
		if err != nil {
			return response.SendErrorResponse(c, fiber.StatusUnauthorized, "refresh_session_not_found")
		}

		// Delete old session (rotation)
		db.DB.Delete(&session)
	}

	// Fetch user and user detail
	var user User
	if err := db.DB.Select("id, email").Where("id = ?", userID).First(&user).Error; err != nil {
		return response.SendErrorResponse(c, fiber.StatusUnauthorized, "user_not_found")
	}
	var userDetail UserDetail
	if err := db.DB.Select("id, first_name, last_name, user_id").Where("user_id = ?", user.ID).First(&userDetail).Error; err != nil {
		return response.SendErrorResponse(c, fiber.StatusUnauthorized, "user_detail_not_found")
	}

	// Generate new access token
	accessTokenExpireMinutes := os.Getenv("ACCESS_TOKEN_EXPIRE_MINUTES")
	accessTokenExpire := 15
	if accessTokenExpireMinutes != "" {
		if v, err := strconv.Atoi(accessTokenExpireMinutes); err == nil {
			accessTokenExpire = v
		}
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Duration(accessTokenExpire) * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     uuid.NewString(),
	})
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "Failed to generate access token")
	}

	// Generate new refresh token
	refreshTokenExpireDays := os.Getenv("REFRESH_TOKEN_EXPIRE_DAYS")
	refreshTokenExpire := 7
	if refreshTokenExpireDays != "" {
		if v, err := strconv.Atoi(refreshTokenExpireDays); err == nil {
			refreshTokenExpire = v
		}
	}
	newRefreshJTI := uuid.NewString()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Duration(refreshTokenExpire) * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     newRefreshJTI,
	})
	newRefreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return response.SendErrorResponse(c, fiber.StatusInternalServerError, "internal_error", "Failed to generate refresh token")
	}

	// Store new session if stateful
	if sessionMode == "jwt_server_stateful" {
		claims := refreshToken.Claims.(jwt.MapClaims)
		issuedAt, _ := claims["iat"].(int64)
		if issuedAt == 0 {
			if f, ok := claims["iat"].(float64); ok {
				issuedAt = int64(f)
			}
		}
		expiresAt, _ := claims["exp"].(int64)
		if expiresAt == 0 {
			if f, ok := claims["exp"].(float64); ok {
				expiresAt = int64(f)
			}
		}
		session := UserSession{
			UserID:     user.ID,
			JTI:        newRefreshJTI,
			CreatedAt:  time.Unix(issuedAt, 0),
			ExpiresAt:  time.Unix(expiresAt, 0),
			LastSeenAt: time.Unix(issuedAt, 0),
		}
		if err := db.DB.Create(&session).Error; err != nil {
			return response.SendErrorResponse(c, fiber.StatusInternalServerError, "db_error", "Failed to create user session")
		}
	}

	// Set new refresh token as HttpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshTokenString,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Expires:  time.Now().Add(time.Duration(refreshTokenExpire) * 24 * time.Hour),
		Secure:   os.Getenv("ENV") == "production", // Only send cookie over HTTPS in production
		Path:     "/",
	})

	// Return success response with new access token
	return response.SendSuccessResponse(c, "Token refreshed successfully", fiber.Map{
		"id":           user.ID,
		"first_name":   userDetail.FirstName,
		"last_name":    userDetail.LastName,
		"email":        user.Email,
		"access_token": accessTokenString,
	})
}
