package user

import (
	"os"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
