package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/sonyarianto/gobete/internal/modules/home"
	"github.com/sonyarianto/gobete/internal/modules/user"
	"github.com/sonyarianto/gobete/internal/systems/http/middleware"

	"os"
	"time"
)

func RegisterRoutes(app *fiber.App) {
	allowedOrigins := "http://localhost:5173" // Default for development
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		allowedOrigins = origins
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins, // or "*" for all origins (not recommended for production)
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowCredentials: true,
	}))

	// Apply rate limiting globally (e.g., 60 requests per minute per IP)
	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 60 * time.Second,
		Next: func(c *fiber.Ctx) bool {
			// Skip rate limiting for versioned /refresh and not versioned /healthz
			return c.Path() == "/v1/refresh" || c.Path() == "/healthz"
		},
	}))

	app.Get("/healthz", HealthCheckHandler)

	// Home route, without version prefix
	app.Get("/", home.HomeHandler)

	// Versioned API v1
	apiV1 := app.Group("/v1")
	RegisterAPIV1Routes(apiV1)
}

// RegisterAPIV1Routes handles all v1 routes
func RegisterAPIV1Routes(api fiber.Router) {
	// Public routes
	api.Get("/", home.HomeHandler)
	api.Post("/login", user.LoginUserHandler)
	api.Post("/users", user.CreateUserHandler)
	api.Post("/refresh", user.RefreshTokenHandler)

	// Protected user routes
	protectedUser := api.Group("/users", middleware.JWTProtected(), middleware.UserSessionCheck())

	// Current user routes
	protectedUser.Get("/me", user.GetCurrentUserHandler)
	protectedUser.Put("/me", user.UpdateCurrentUserHandler)
	protectedUser.Put("/me/password", user.ChangePasswordHandler)
	protectedUser.Delete("/me", user.DeleteCurrentUserHandler)

	// Admin-only routes
	adminUsers := protectedUser.Group("/", middleware.AdminOnly())
	adminUsers.Get("/", user.ListUsersHandler)
	adminUsers.Get("/:id", user.GetUserByIDHandler)
	adminUsers.Put("/:id", user.UpdateUserByIDHandler)
	adminUsers.Delete("/:id", user.DeleteUserByIDHandler)

	// Logout (protected)
	api.Post("/logout", user.LogoutUserHandler)
}
