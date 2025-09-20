package middleware

import (
	"github.com/sonyarianto/gobete/internal/modules/user"
	"github.com/sonyarianto/gobete/internal/systems/db"
	"github.com/sonyarianto/gobete/internal/systems/response"

	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return response.SendErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", "Unauthorized access - invalid or missing token.")
		}

		accessTokenString := strings.TrimPrefix(authHeader, "Bearer ")
		accessToken, err := jwt.Parse(accessTokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !accessToken.Valid {
			return response.SendErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", "Unauthorized access - invalid or missing token.")
		}

		c.Locals("user", accessToken)
		return c.Next()
	}
}

func UserSessionCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only check session if SESSION_MODE is "jwt_server_stateful"
		if os.Getenv("SESSION_MODE") != "jwt_server_stateful" {
			return c.Next()
		}

		// Try to get refresh_token from cookie
		refreshToken := c.Cookies("refresh_token")
		if refreshToken == "" {
			return response.SendErrorResponse(c, fiber.StatusUnauthorized, "session_expired")
		}

		token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			return response.SendErrorResponse(c, fiber.StatusUnauthorized, "session_expired")
		}

		claims := token.Claims.(jwt.MapClaims)
		if !isValidUserSession(claims) {
			return response.SendErrorResponse(c, fiber.StatusUnauthorized, "session_expired")
		}

		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	// TODO: Implement admin check logic
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func isValidUserSession(claims jwt.MapClaims) bool {
	var userID string
	switch v := claims["user_id"].(type) {
	case string:
		userID = v
	case float64:
		userID = fmt.Sprintf("%.0f", v) // JWT numbers are float64
	default:
		return false
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return false
	}

	var sessionCount int64
	db.DB.Model(&user.UserSession{}).
		Where("user_id = ? AND jti = ? AND expires_at > ?", userID, jti, time.Now()).
		Count(&sessionCount)

	return sessionCount > 0
}
