package utility

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Map validation tags to friendly messages
func ValidationLoginMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return field + " is required."
	case "email":
		return "Invalid email address format."
	default:
		return field + " is invalid."
	}
}

// Format validation errors using JSON field names and custom messages
func FormatValidationErrors(structVal any, validationErrors validator.ValidationErrors) map[string][]map[string]any {
	errors := make(map[string][]map[string]any)
	val := reflect.ValueOf(structVal)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	typ := val.Type()
	for _, fieldErr := range validationErrors {
		field, ok := typ.FieldByName(fieldErr.Field())
		jsonTag := fieldErr.Field()
		if ok {
			tag := field.Tag.Get("json")
			if tag != "" && tag != "-" {
				jsonTag = tag
			}
		}
		errorObj := map[string]any{
			"rule":    fieldErr.Tag(),
			"param":   fieldErr.Param(),
			"message": ValidationLoginMessage(jsonTag, fieldErr.Tag(), fieldErr.Param()),
		}
		errors[jsonTag] = append(errors[jsonTag], errorObj)
	}
	return errors
}
