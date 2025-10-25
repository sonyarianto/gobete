package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sonyarianto/gobete/internal/systems/db"
	"github.com/sonyarianto/gobete/internal/systems/response"

	"os"
	"time"
)

func LogoutUserHandler(c *fiber.Ctx) error {
	// If session mode is stateful, delete the session from DB, actually the refresh token JTI
	if os.Getenv("SESSION_MODE") == "jwt_server_stateful" {
		refreshTokenString := c.Cookies("refresh_token")
		if refreshTokenString != "" {
			token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (any, error) {
				return jwtSecret, nil
			})
			if err == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					if jti, ok := claims["jti"].(string); ok {
						// Ignore DB errors for idempotency
						_ = db.DB.Where("jti = ?", jti).Delete(&UserSession{}).Error
					}
				}
			}
			// If token is invalid, just continue (do not return error)
		}
	}

	// Clear the refresh token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Expire immediately
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   os.Getenv("ENV") == "production",
		Path:     "/",
	})

	// Always return the same message
	return response.SendSuccessResponse(c, "User logged out successfully", nil)
}
