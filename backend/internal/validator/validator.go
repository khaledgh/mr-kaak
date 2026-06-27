// Package validator wraps go-playground/validator with project-specific rules
// (Canadian postal code & province, E.164 phone, RBAC role) and converts
// validation failures into structured, client-friendly field errors that slot
// into the response envelope's error details.
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	govalidator "github.com/go-playground/validator/v10"
	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/pkg/canada"
	"github.com/nyaruka/phonenumbers"
)

// slugRe matches URL-safe slugs: lowercase alphanumerics separated by single
// hyphens (e.g. "kaak-cheese-basterma").
var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Validator validates request DTOs.
type Validator struct {
	v *govalidator.Validate
}

// FieldError describes a single failed field, suitable for JSON details.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError aggregates field errors from one validation pass.
type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.Field + ": " + f.Message
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// New builds a Validator with custom rules registered.
func New() *Validator {
	v := govalidator.New(govalidator.WithRequiredStructEnabled())

	// Report JSON tag names in errors instead of Go struct field names.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	_ = v.RegisterValidation("ca_postal", func(fl govalidator.FieldLevel) bool {
		return canada.ValidPostal(fl.Field().String())
	})
	_ = v.RegisterValidation("ca_province", func(fl govalidator.FieldLevel) bool {
		return canada.ValidProvince(fl.Field().String())
	})
	_ = v.RegisterValidation("e164ca", func(fl govalidator.FieldLevel) bool {
		_, err := NormalizePhoneCA(fl.Field().String())
		return err == nil
	})
	_ = v.RegisterValidation("role", func(fl govalidator.FieldLevel) bool {
		return models.Role(fl.Field().String()).Valid()
	})
	_ = v.RegisterValidation("slug", func(fl govalidator.FieldLevel) bool {
		return slugRe.MatchString(fl.Field().String())
	})

	return &Validator{v: v}
}

// Struct validates s and returns a *ValidationError (or nil if valid).
func (val *Validator) Struct(s any) error {
	err := val.v.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *govalidator.InvalidValidationError
	if errors.As(err, &invalid) {
		return err // programmer error (nil / non-struct passed)
	}

	var verrs govalidator.ValidationErrors
	if errors.As(err, &verrs) {
		out := &ValidationError{Fields: make([]FieldError, 0, len(verrs))}
		for _, fe := range verrs {
			out.Fields = append(out.Fields, FieldError{
				Field:   fe.Field(),
				Message: messageFor(fe),
			})
		}
		return out
	}
	return err
}

func messageFor(fe govalidator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "ca_postal":
		return "must be a valid Canadian postal code (e.g. A1A 1A1)"
	case "ca_province":
		return "must be a valid Canadian province/territory code"
	case "e164ca":
		return "must be a valid Canadian phone number"
	case "role":
		return "must be a valid role"
	case "oneof":
		return "must be one of: " + fe.Param()
	case "slug":
		return "must be a lowercase slug (letters, numbers, hyphens)"
	case "url":
		return "must be a valid URL"
	default:
		return "is invalid"
	}
}

// NormalizePhoneCA parses a phone number in the Canadian region and returns it
// in E.164 form (+1XXXXXXXXXX). Empty input is rejected; callers that allow an
// optional phone should guard for "" before calling.
func NormalizePhoneCA(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty phone")
	}
	num, err := phonenumbers.Parse(raw, "CA")
	if err != nil {
		return "", fmt.Errorf("parse phone: %w", err)
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", errors.New("invalid phone number")
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
