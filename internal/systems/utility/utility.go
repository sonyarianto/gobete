package utility

import (
	"reflect"
	"strings"
)

// Helper to get JSON tag from struct type and field name
func GetJSONFieldName(obj any, fieldName string) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if f, ok := t.FieldByName(fieldName); ok {
		tag := f.Tag.Get("json")
		if tag != "" && tag != "-" {
			return strings.Split(tag, ",")[0]
		}
	}
	return fieldName
}
