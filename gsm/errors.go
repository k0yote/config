package gsm

import (
	"errors"
	"fmt"
)

var (
	// ErrSecretNotFound is returned when a secret is not found in Secret Manager or environment variables
	// and no default value is provided.
	ErrSecretNotFound = errors.New("secret not found")

	// ErrInvalidTarget is returned when the target for Load() is not a pointer to a struct.
	ErrInvalidTarget = errors.New("target must be a pointer to a struct")

	// ErrRequiredFieldMissing is returned when a required field (marked with 'required' tag) has no value.
	ErrRequiredFieldMissing = errors.New("required field is missing")

	// ErrInvalidFormat is returned when the secret reference format is invalid.
	ErrInvalidFormat = errors.New("invalid secret reference format")

	// ErrUnsupportedType is returned when trying to set a value to an unsupported field type.
	ErrUnsupportedType = errors.New("unsupported field type")
)

// SecretNotFoundError wraps ErrSecretNotFound with additional context.
type SecretNotFoundError struct {
	SecretName string
}

func (e *SecretNotFoundError) Error() string {
	return fmt.Sprintf("secret not found: %s", e.SecretName)
}

func (e *SecretNotFoundError) Unwrap() error {
	return ErrSecretNotFound
}

// RequiredFieldError wraps ErrRequiredFieldMissing with field information.
type RequiredFieldError struct {
	FieldName  string
	SecretName string
}

func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("required field '%s' (secret: %s) is missing", e.FieldName, e.SecretName)
}

func (e *RequiredFieldError) Unwrap() error {
	return ErrRequiredFieldMissing
}

// InvalidFormatError wraps ErrInvalidFormat with the invalid value.
type InvalidFormatError struct {
	Value  string
	Reason string
}

func (e *InvalidFormatError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("invalid format: %s (%s)", e.Value, e.Reason)
	}
	return fmt.Sprintf("invalid format: %s", e.Value)
}

func (e *InvalidFormatError) Unwrap() error {
	return ErrInvalidFormat
}

// UnsupportedTypeError wraps ErrUnsupportedType with type information.
type UnsupportedTypeError struct {
	FieldName string
	TypeName  string
}

func (e *UnsupportedTypeError) Error() string {
	return fmt.Sprintf("unsupported type for field '%s': %s", e.FieldName, e.TypeName)
}

func (e *UnsupportedTypeError) Unwrap() error {
	return ErrUnsupportedType
}
