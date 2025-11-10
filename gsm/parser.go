package gsm

import (
	"strings"
)

const (
	// SecretPrefix is the prefix used to identify secret references in configuration values.
	SecretPrefix = "sm://"

	// DefaultSeparator separates the secret name from the default value.
	DefaultSeparator = "||"
)

// SecretRef represents a parsed secret reference with its components.
type SecretRef struct {
	// SecretName is the name of the environment variable or secret in Secret Manager.
	SecretName string

	// DefaultValue is the fallback value if the secret is not found.
	// Empty string if no default was specified.
	DefaultValue string

	// HasDefault indicates whether a default value was provided.
	HasDefault bool

	// IsSecretRef indicates whether the original value had the "sm://" prefix.
	IsSecretRef bool
}

// Parse parses a value that may contain a secret reference in the format:
// "sm://SECRET_NAME||default_value" or "sm://SECRET_NAME"
//
// If the value doesn't start with "sm://", it returns a SecretRef with IsSecretRef=false
// and the original value as DefaultValue.
//
// Examples:
//   - "sm://API_KEY||default" -> SecretRef{SecretName: "API_KEY", DefaultValue: "default", HasDefault: true, IsSecretRef: true}
//   - "sm://API_KEY" -> SecretRef{SecretName: "API_KEY", HasDefault: false, IsSecretRef: true}
//   - "plain_value" -> SecretRef{DefaultValue: "plain_value", HasDefault: true, IsSecretRef: false}
func Parse(value string) SecretRef {
	// If it doesn't start with sm://, treat it as a plain value
	after, found := strings.CutPrefix(value, SecretPrefix)
	if !found {
		return SecretRef{
			DefaultValue: value,
			HasDefault:   true,
			IsSecretRef:  false,
		}
	}

	// Use the value after the prefix
	value = after

	// Split by the separator
	parts := strings.SplitN(value, DefaultSeparator, 2)

	ref := SecretRef{
		SecretName:  strings.TrimSpace(parts[0]),
		IsSecretRef: true,
	}

	// If there's a default value part
	if len(parts) == 2 {
		ref.DefaultValue = parts[1] // Don't trim spaces from default value
		ref.HasDefault = true
	}

	return ref
}

// ParseSlice parses a slice of values, each of which may contain secret references.
// This is useful for configuration values that are arrays.
func ParseSlice(values []string) []SecretRef {
	refs := make([]SecretRef, len(values))
	for i, v := range values {
		refs[i] = Parse(v)
	}
	return refs
}

// IsSecretReference checks if a value is a secret reference (starts with "sm://").
func IsSecretReference(value string) bool {
	return strings.HasPrefix(value, SecretPrefix)
}
