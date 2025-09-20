package error

var ErrorMessages = map[string]string{
	"not_found":                    "Resource not found.",
	"invalid_payload":              "Invalid request payload.",
	"unauthorized":                 "Invalid email or password.",
	"invalid_credentials":          "Invalid email or password.",
	"validation_error":             "Validation error.",
	"user_exists":                  "User already exists.",
	"internal_error":               "Internal server error.",
	"db_error":                     "Database error.",
	"bad_request":                  "Bad request.",
	"record_not_found":             "Record not found.",
	"session_expired":              "Session has expired. Please log in again.",
	"no_refresh_token":             "No refresh token provided.",
	"invalid_refresh_token":        "Invalid refresh token.",
	"invalid_refresh_token_claims": "Invalid refresh token claims.",
	"invalid_user_id":              "Invalid user ID.",
	"invalid_jti":                  "Invalid token identifier (jti).",
	"refresh_session_not_found":    "Refresh session not found.",
	"user_not_found":               "User not found.",
	"user_detail_not_found":        "User detail not found.",
	// Add more error codes and messages as needed
}
