package user

import (
	"fmt"
)

// Map validation tags to friendly messages
func ValidationLoginMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required.", field)
	case "email":
		return "Invalid email address format."
	// case "min":
	// 	return fmt.Sprintf("%s must be at least %s characters long.", field, param)
	// case "max":
	// 	return fmt.Sprintf("%s must not exceed %s characters.", field, param)
	// Add more tag cases as needed
	default:
		return fmt.Sprintf("%s is invalid.", field)
	}
}
