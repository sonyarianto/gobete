package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model        // Includes ID, CreatedAt, UpdatedAt, DeletedAt
	Email      string `json:"email" validate:"required,email" gorm:"unique"`
	Password   string `json:"password" validate:"required,min=8"`
	// Add other fields as needed
}

type UserDetail struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"uniqueIndex"` // Ensure one-to-one relationship
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// Add more fields as needed
}

type UserSession struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"uniqueIndex"` // Ensure one-to-one relationship
	JTI        string    `json:"jti" gorm:"uniqueIndex"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}
