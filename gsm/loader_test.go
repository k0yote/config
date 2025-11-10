package gsm

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoaderLoad(t *testing.T) {
	ctx := context.Background()

	t.Run("load string fields", func(t *testing.T) {
		type Config struct {
			Field1 string `gsm:"FIELD1,default=default1"`
			Field2 string `gsm:"FIELD2,default=default2"`
		}

		os.Setenv("FIELD1", "value1")
		defer os.Unsetenv("FIELD1")

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "value1", cfg.Field1)
		assert.Equal(t, "default2", cfg.Field2)
	})

	t.Run("load int fields", func(t *testing.T) {
		type Config struct {
			Port     int   `gsm:"PORT,default=8080"`
			MaxConns int64 `gsm:"MAX_CONNS,default=100"`
		}

		os.Setenv("PORT", "3000")
		defer os.Unsetenv("PORT")

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.Equal(t, 3000, cfg.Port)
		assert.Equal(t, int64(100), cfg.MaxConns)
	})

	t.Run("load bool fields", func(t *testing.T) {
		type Config struct {
			Debug   bool `gsm:"DEBUG,default=false"`
			Verbose bool `gsm:"VERBOSE,default=true"`
		}

		os.Setenv("DEBUG", "true")
		defer os.Unsetenv("DEBUG")

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.True(t, cfg.Debug)
		assert.True(t, cfg.Verbose)
	})

	t.Run("load float fields", func(t *testing.T) {
		type Config struct {
			Rate    float64 `gsm:"RATE,default=1.5"`
			Timeout float32 `gsm:"TIMEOUT,default=30.0"`
		}

		os.Setenv("RATE", "2.5")
		defer os.Unsetenv("RATE")

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.Equal(t, 2.5, cfg.Rate)
		assert.Equal(t, float32(30.0), cfg.Timeout)
	})

	t.Run("load slice fields", func(t *testing.T) {
		type Config struct {
			Hosts []string `gsm:"HOSTS,default=localhost,127.0.0.1"`
			Tags  []string `gsm:"TAGS"`
		}

		os.Setenv("HOSTS", `["host1.com", "host2.com"]`)
		os.Setenv("TAGS", "tag1,tag2,tag3")
		defer os.Unsetenv("HOSTS")
		defer os.Unsetenv("TAGS")

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.Equal(t, []string{"host1.com", "host2.com"}, cfg.Hosts)
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, cfg.Tags)
	})

	t.Run("required field present", func(t *testing.T) {
		type Config struct {
			APIKey string `gsm:"API_KEY,required"`
		}

		os.Setenv("API_KEY", "secret_key")
		defer os.Unsetenv("API_KEY")

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "secret_key", cfg.APIKey)
	})

	t.Run("required field missing", func(t *testing.T) {
		type Config struct {
			APIKey string `gsm:"API_KEY,required"`
		}

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.Error(t, err)
		var reqErr *RequiredFieldError
		require.ErrorAs(t, err, &reqErr)
		assert.Equal(t, "APIKey", reqErr.FieldName)
	})

	t.Run("skip ignored field", func(t *testing.T) {
		type Config struct {
			Field1 string `gsm:"FIELD1,default=value1"`
			Field2 string `gsm:"-"`
		}

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "value1", cfg.Field1)
		assert.Equal(t, "", cfg.Field2) // Should remain empty
	})

	t.Run("skip field without tag", func(t *testing.T) {
		type Config struct {
			Field1 string `gsm:"FIELD1,default=value1"`
			Field2 string // No tag
		}

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "value1", cfg.Field1)
		assert.Equal(t, "", cfg.Field2) // Should remain empty
	})

	t.Run("with env prefix", func(t *testing.T) {
		type Config struct {
			Field1 string `gsm:"FIELD1,default=default1"`
		}

		os.Setenv("APP_FIELD1", "prefixed_value")
		defer os.Unsetenv("APP_FIELD1")

		loader := NewLoader(nil, WithEnvPrefix("APP_"), WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.NoError(t, err)
		assert.Equal(t, "prefixed_value", cfg.Field1)
	})

	t.Run("invalid target - not pointer", func(t *testing.T) {
		type Config struct {
			Field string `gsm:"FIELD"`
		}

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, cfg) // Not a pointer

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTarget)
	})

	t.Run("invalid target - pointer to non-struct", func(t *testing.T) {
		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var str string
		err := loader.Load(ctx, &str)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTarget)
	})

	t.Run("invalid int value", func(t *testing.T) {
		type Config struct {
			Port int `gsm:"PORT,default=not_a_number,required"`
		}

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.Error(t, err)
		var reqErr *RequiredFieldError
		require.ErrorAs(t, err, &reqErr)
	})

	t.Run("invalid bool value", func(t *testing.T) {
		type Config struct {
			Debug bool `gsm:"DEBUG,default=not_a_bool,required"`
		}

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		require.Error(t, err)
		var reqErr *RequiredFieldError
		require.ErrorAs(t, err, &reqErr)
	})

	t.Run("optional field not required", func(t *testing.T) {
		type Config struct {
			Optional string `gsm:"OPTIONAL"`
		}

		loader := NewLoader(nil, WithSecretManagerEnabled(false))
		var cfg Config
		err := loader.Load(ctx, &cfg)

		// Should not error, field remains empty
		require.NoError(t, err)
		assert.Equal(t, "", cfg.Optional)
	})
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected tagInfo
	}{
		{
			name: "simple secret name",
			tag:  "SECRET_NAME",
			expected: tagInfo{
				secretName: "SECRET_NAME",
				hasDefault: false,
				required:   false,
			},
		},
		{
			name: "with default",
			tag:  "SECRET_NAME,default=default_value",
			expected: tagInfo{
				secretName:   "SECRET_NAME",
				defaultValue: "default_value",
				hasDefault:   true,
				required:     false,
			},
		},
		{
			name: "with required",
			tag:  "SECRET_NAME,required",
			expected: tagInfo{
				secretName: "SECRET_NAME",
				hasDefault: false,
				required:   true,
			},
		},
		{
			name: "with default and required",
			tag:  "SECRET_NAME,default=value,required",
			expected: tagInfo{
				secretName:   "SECRET_NAME",
				defaultValue: "value",
				hasDefault:   true,
				required:     true,
			},
		},
		{
			name: "with spaces",
			tag:  " SECRET_NAME , default=value , required ",
			expected: tagInfo{
				secretName:   "SECRET_NAME",
				defaultValue: "value",
				hasDefault:   true,
				required:     true,
			},
		},
		{
			name: "default with comma",
			tag:  "SECRET_NAME,default=value1,value2",
			expected: tagInfo{
				secretName:   "SECRET_NAME",
				defaultValue: "value1",
				hasDefault:   true,
				required:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTag(tt.tag)
			assert.Equal(t, tt.expected.secretName, result.secretName)
			assert.Equal(t, tt.expected.defaultValue, result.defaultValue)
			assert.Equal(t, tt.expected.hasDefault, result.hasDefault)
			assert.Equal(t, tt.expected.required, result.required)
		})
	}
}
