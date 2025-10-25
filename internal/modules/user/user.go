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
	"strings"
	"time"
)

var validate = validator.New(validator.WithRequiredStructEnabled())
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func CreateUserHandler(c *fiber.Ctx) error {
	var req CreateUserRequest

	req = *c.Locals("body").(*CreateUserRequest)

	// Validate input
	if err := validate.Struct(&req); err != nil {
		// Collect validation errors in a map with detailed info
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]map[string]string)
			for _, fieldErr := range validationErrors {
				fieldName := strings.ToLower(fieldErr.Field())
				errors[fieldName] = map[string]string{
					"tag":     fieldErr.Tag(),
					"param":   fieldErr.Param(),
					"message": fieldErr.Error(),
				}
			}
			return response.SendErrorResponse(c, fiber.StatusBadRequest, "validation_error", errors)
		}
		// Fallback for other errors
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

func LoginUserHandler(c *fiber.Ctx) error {
	var req LoginRequest

	req = *c.Locals("body").(*LoginRequest)

	// Validate input
	if err := validate.Struct(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string][]map[string]any)
			for _, fieldErr := range validationErrors {
				// field := strings.ToLower(fieldErr.Field()) // Normalize to lowercase JSON field names
				jsonField := utility.GetJSONFieldName(req, fieldErr.Field())
				errorObj := map[string]any{
					"rule":    fieldErr.Tag(),
					"param":   fieldErr.Param(),
					"message": ValidationLoginMessage(jsonField, fieldErr.Tag(), fieldErr.Param()),
				}
				errors[jsonField] = append(errors[jsonField], errorObj)
			}
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

func UpdateCurrentUserHandler(c *fiber.Ctx) error {
	// TODO: Implement update current user profile logic
	return response.SendSuccessResponse(c, "Update current user - not implemented yet", nil)
}

func ChangePasswordHandler(c *fiber.Ctx) error {
	// TODO: Implement change password logic
	return response.SendSuccessResponse(c, "Change password - not implemented yet", nil)
}

func DeleteCurrentUserHandler(c *fiber.Ctx) error {
	// TODO: Implement delete current user logic
	return response.SendSuccessResponse(c, "Delete current user - not implemented yet", nil)
}

func ListUsersHandler(c *fiber.Ctx) error {
	// TODO: Implement list users logic (admin only)
	return response.SendSuccessResponse(c, "List users - not implemented yet", nil)
}

func GetUserByIDHandler(c *fiber.Ctx) error {
	// TODO: Implement get user by ID logic (admin only)
	return response.SendSuccessResponse(c, "Get user by ID - not implemented yet", nil)
}

func DeleteUserByIDHandler(c *fiber.Ctx) error {
	// TODO: Implement delete user by ID logic (admin only)
	return response.SendSuccessResponse(c, "Delete user by ID - not implemented yet", nil)
}

func UpdateUserByIDHandler(c *fiber.Ctx) error {
	// TODO: Implement update user by ID logic (admin only)
	return response.SendSuccessResponse(c, "Update user by ID - not implemented yet", nil)
}
