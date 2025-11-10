package gsm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolverResolve(t *testing.T) {
	ctx := context.Background()

	t.Run("resolve from environment variable", func(t *testing.T) {
		// Set up environment variable
		os.Setenv("TEST_KEY", "test_value")
		defer os.Unsetenv("TEST_KEY")

		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		value, err := resolver.Resolve(ctx, "sm://TEST_KEY||default")

		require.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("resolve with default when env var not found", func(t *testing.T) {
		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		value, err := resolver.Resolve(ctx, "sm://NON_EXISTENT||default_value")

		require.NoError(t, err)
		assert.Equal(t, "default_value", value)
	})

	t.Run("error when no default and not found", func(t *testing.T) {
		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		_, err := resolver.Resolve(ctx, "sm://NON_EXISTENT")

		require.Error(t, err)
		var notFoundErr *SecretNotFoundError
		assert.ErrorAs(t, err, &notFoundErr)
	})

	t.Run("resolve plain value", func(t *testing.T) {
		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		value, err := resolver.Resolve(ctx, "plain_value")

		require.NoError(t, err)
		assert.Equal(t, "plain_value", value)
	})

	t.Run("env prefix", func(t *testing.T) {
		os.Setenv("APP_TEST_KEY", "prefixed_value")
		defer os.Unsetenv("APP_TEST_KEY")

		resolver := NewResolver(nil, WithEnvPrefix("APP_"), WithSecretManagerEnabled(false))
		value, err := resolver.Resolve(ctx, "sm://TEST_KEY||default")

		require.NoError(t, err)
		assert.Equal(t, "prefixed_value", value)
	})

	t.Run("empty env var uses default", func(t *testing.T) {
		os.Setenv("EMPTY_KEY", "")
		defer os.Unsetenv("EMPTY_KEY")

		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		value, err := resolver.Resolve(ctx, "sm://EMPTY_KEY||default_value")

		require.NoError(t, err)
		assert.Equal(t, "default_value", value)
	})
}

func TestResolverResolveSlice(t *testing.T) {
	ctx := context.Background()

	t.Run("resolve JSON array from env", func(t *testing.T) {
		os.Setenv("ARRAY_KEY", `["value1", "value2", "value3"]`)
		defer os.Unsetenv("ARRAY_KEY")

		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		values, err := resolver.ResolveSlice(ctx, []string{"sm://ARRAY_KEY"})

		require.NoError(t, err)
		assert.Equal(t, []string{"value1", "value2", "value3"}, values)
	})

	t.Run("resolve CSV from env", func(t *testing.T) {
		os.Setenv("CSV_KEY", "value1,value2,value3")
		defer os.Unsetenv("CSV_KEY")

		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		values, err := resolver.ResolveSlice(ctx, []string{"sm://CSV_KEY"})

		require.NoError(t, err)
		assert.Equal(t, []string{"value1", "value2", "value3"}, values)
	})

	t.Run("resolve CSV with spaces", func(t *testing.T) {
		os.Setenv("CSV_SPACES", "value1 , value2 , value3")
		defer os.Unsetenv("CSV_SPACES")

		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		values, err := resolver.ResolveSlice(ctx, []string{"sm://CSV_SPACES"})

		require.NoError(t, err)
		assert.Equal(t, []string{"value1", "value2", "value3"}, values)
	})

	t.Run("resolve with default array", func(t *testing.T) {
		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		values, err := resolver.ResolveSlice(ctx, []string{"sm://NON_EXISTENT||default1,default2"})

		require.NoError(t, err)
		assert.Equal(t, []string{"default1", "default2"}, values)
	})

	t.Run("resolve multiple secret refs", func(t *testing.T) {
		os.Setenv("KEY1", "value1")
		os.Setenv("KEY2", "value2")
		defer os.Unsetenv("KEY1")
		defer os.Unsetenv("KEY2")

		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		values, err := resolver.ResolveSlice(ctx, []string{"sm://KEY1", "sm://KEY2"})

		require.NoError(t, err)
		assert.Equal(t, []string{"value1", "value2"}, values)
	})

	t.Run("empty slice", func(t *testing.T) {
		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		values, err := resolver.ResolveSlice(ctx, []string{})

		require.NoError(t, err)
		assert.Equal(t, []string{}, values)
	})

	t.Run("single value", func(t *testing.T) {
		os.Setenv("SINGLE_KEY", "single_value")
		defer os.Unsetenv("SINGLE_KEY")

		resolver := NewResolver(nil, WithSecretManagerEnabled(false))
		values, err := resolver.ResolveSlice(ctx, []string{"sm://SINGLE_KEY"})

		require.NoError(t, err)
		assert.Equal(t, []string{"single_value"}, values)
	})
}

func TestParseArrayValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name:     "JSON array",
			input:    `["value1", "value2", "value3"]`,
			expected: []string{"value1", "value2", "value3"},
			wantErr:  false,
		},
		{
			name:     "CSV",
			input:    "value1,value2,value3",
			expected: []string{"value1", "value2", "value3"},
			wantErr:  false,
		},
		{
			name:     "CSV with spaces",
			input:    "value1 , value2 , value3",
			expected: []string{"value1", "value2", "value3"},
			wantErr:  false,
		},
		{
			name:     "single value",
			input:    "single_value",
			expected: []string{"single_value"},
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "invalid JSON",
			input:    `["value1", "value2"`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseArrayValue(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
