package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator to provide Indonesian error messages
// and JSON-tag-based field names.
type Validator struct {
	v *validator.Validate
}

// New creates a ready-to-use Validator instance.
func New() *Validator {
	return &Validator{v: validator.New()}
}

// Validate validates the given struct and returns a map of field → error message.
// Keys use the JSON tag name of each field. Returns nil when there are no errors.
func (vl *Validator) Validate(i interface{}) map[string]string {
	err := vl.v.Struct(i)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{"error": "tidak valid"}
	}

	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	result := make(map[string]string, len(validationErrors))
	for _, fe := range validationErrors {
		key := jsonFieldName(t, fe.Field())
		result[key] = humanMessage(fe)
	}
	return result
}

// jsonFieldName resolves the JSON tag for a field name within the given type.
// Falls back to the lowercase Go field name when no json tag is set.
func jsonFieldName(t reflect.Type, fieldName string) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return strings.ToLower(fieldName)
	}

	sf, ok := t.FieldByName(fieldName)
	if !ok {
		return strings.ToLower(fieldName)
	}

	tag := sf.Tag.Get("json")
	if tag == "" || tag == "-" {
		return strings.ToLower(fieldName)
	}

	// json tag may be "name,omitempty" — take only the name part
	return strings.SplitN(tag, ",", 2)[0]
}

// humanMessage converts a validator.FieldError into a human-readable
// Indonesian error message.
func humanMessage(fe validator.FieldError) string {
	param := fe.Param()
	kind := fe.Kind()

	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", param)
	case "uuid4":
		return "format UUID tidak valid"
	case "len":
		return fmt.Sprintf("harus tepat %s karakter", param)
	case "numeric":
		return "harus berupa angka"
	case "min":
		if kind == reflect.String {
			return fmt.Sprintf("minimal %s karakter", param)
		}
		return fmt.Sprintf("minimal %s", param)
	case "max":
		if kind == reflect.String {
			return fmt.Sprintf("maksimal %s karakter", param)
		}
		return fmt.Sprintf("maksimal %s", param)
	default:
		return "tidak valid"
	}
}
