package gsm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Resolver resolves configuration values from environment variables, Secret Manager, or defaults.
type Resolver struct {
	client               *Client
	secretManagerEnabled bool
	envPrefix            string
}

// ResolverOption is a functional option for configuring a Resolver.
type ResolverOption func(*Resolver)

// WithSecretManagerEnabled controls whether Secret Manager lookups are performed.
// If false, only environment variables and defaults are used.
func WithSecretManagerEnabled(enabled bool) ResolverOption {
	return func(r *Resolver) {
		r.secretManagerEnabled = enabled
	}
}

// WithEnvPrefix sets a prefix that will be added to all environment variable lookups.
// For example, with prefix "APP_", looking up "DB_HOST" will check "APP_DB_HOST".
func WithEnvPrefix(prefix string) ResolverOption {
	return func(r *Resolver) {
		r.envPrefix = prefix
	}
}

// NewResolver creates a new Resolver with the given client and options.
// The client can be nil if Secret Manager is not used.
func NewResolver(client *Client, opts ...ResolverOption) *Resolver {
	r := &Resolver{
		client:               client,
		secretManagerEnabled: client != nil,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Resolve resolves a single value using the priority: env var -> Secret Manager -> default.
//
// The value parameter can be:
//   - A secret reference: "sm://SECRET_NAME||default_value"
//   - A plain value: "some_value" (returned as-is)
//
// Returns the resolved value or an error if the value couldn't be resolved and no default exists.
func (r *Resolver) Resolve(ctx context.Context, value string) (string, error) {
	ref := Parse(value)

	// If it's not a secret reference, return the value as-is
	if !ref.IsSecretRef {
		return ref.DefaultValue, nil
	}

	// Priority 1: Check environment variable
	envKey := r.envPrefix + ref.SecretName
	if envValue, exists := os.LookupEnv(envKey); exists && envValue != "" {
		return envValue, nil
	}

	// Priority 2: Check Secret Manager (if enabled and client available)
	if r.secretManagerEnabled && r.client != nil {
		smValue, err := r.client.GetSecret(ctx, ref.SecretName)
		if err == nil {
			return smValue, nil
		}
		// If Secret Manager returns an error, continue to default (don't fail immediately)
	}

	// Priority 3: Use default value
	if ref.HasDefault {
		return ref.DefaultValue, nil
	}

	// No value found and no default provided
	return "", &SecretNotFoundError{SecretName: ref.SecretName}
}

// ResolveSlice resolves a slice of values, where the environment variable might contain
// a JSON array or comma-separated values.
//
// If the environment variable contains a JSON array (starts with '['), it will be parsed.
// Otherwise, if it contains commas, it will be split by commas.
// If it's a single value, it will be returned as a single-element slice.
func (r *Resolver) ResolveSlice(ctx context.Context, values []string) ([]string, error) {
	// If the slice is empty, return empty
	if len(values) == 0 {
		return []string{}, nil
	}

	// If we have a single secret reference, try to resolve it as an array source
	if len(values) == 1 && IsSecretReference(values[0]) {
		ref := Parse(values[0])

		// Priority 1: Check environment variable
		envKey := r.envPrefix + ref.SecretName
		if envValue, exists := os.LookupEnv(envKey); exists && envValue != "" {
			return parseArrayValue(envValue)
		}

		// Priority 2: Check Secret Manager (if enabled)
		if r.secretManagerEnabled && r.client != nil {
			smValue, err := r.client.GetSecret(ctx, ref.SecretName)
			if err == nil {
				return parseArrayValue(smValue)
			}
		}

		// Priority 3: Use default value if available
		if ref.HasDefault {
			return parseArrayValue(ref.DefaultValue)
		}

		return nil, &SecretNotFoundError{SecretName: ref.SecretName}
	}

	// If we have multiple values, resolve each one individually
	result := make([]string, 0, len(values))
	for _, v := range values {
		resolved, err := r.Resolve(ctx, v)
		if err != nil {
			return nil, err
		}
		result = append(result, resolved)
	}

	return result, nil
}

// parseArrayValue parses a value that might be a JSON array or comma-separated values.
// Examples:
//   - `["value1", "value2"]` -> ["value1", "value2"]
//   - `value1,value2,value3` -> ["value1", "value2", "value3"]
//   - `single_value` -> ["single_value"]
func parseArrayValue(value string) ([]string, error) {
	value = strings.TrimSpace(value)

	// Try to parse as JSON array
	if strings.HasPrefix(value, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(value), &arr); err != nil {
			return nil, fmt.Errorf("failed to parse JSON array: %w", err)
		}
		return arr, nil
	}

	// Check if it's comma-separated
	if strings.Contains(value, ",") {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, nil
	}

	// Single value
	if value != "" {
		return []string{value}, nil
	}

	return []string{}, nil
}
