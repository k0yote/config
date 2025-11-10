package gsm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected SecretRef
	}{
		{
			name:  "secret reference with default",
			input: "sm://API_KEY||default_key",
			expected: SecretRef{
				SecretName:   "API_KEY",
				DefaultValue: "default_key",
				HasDefault:   true,
				IsSecretRef:  true,
			},
		},
		{
			name:  "secret reference without default",
			input: "sm://API_KEY",
			expected: SecretRef{
				SecretName:   "API_KEY",
				DefaultValue: "",
				HasDefault:   false,
				IsSecretRef:  true,
			},
		},
		{
			name:  "plain value",
			input: "plain_value",
			expected: SecretRef{
				SecretName:   "",
				DefaultValue: "plain_value",
				HasDefault:   true,
				IsSecretRef:  false,
			},
		},
		{
			name:  "secret reference with empty default",
			input: "sm://API_KEY||",
			expected: SecretRef{
				SecretName:   "API_KEY",
				DefaultValue: "",
				HasDefault:   true,
				IsSecretRef:  true,
			},
		},
		{
			name:  "secret reference with spaces",
			input: "sm:// API_KEY ||default",
			expected: SecretRef{
				SecretName:   "API_KEY",
				DefaultValue: "default",
				HasDefault:   true,
				IsSecretRef:  true,
			},
		},
		{
			name:  "default value with spaces",
			input: "sm://API_KEY|| localhost:5432 ",
			expected: SecretRef{
				SecretName:   "API_KEY",
				DefaultValue: " localhost:5432 ",
				HasDefault:   true,
				IsSecretRef:  true,
			},
		},
		{
			name:  "default value with separator",
			input: "sm://API_KEY||value||with||separator",
			expected: SecretRef{
				SecretName:   "API_KEY",
				DefaultValue: "value||with||separator",
				HasDefault:   true,
				IsSecretRef:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			assert.Equal(t, tt.expected.SecretName, result.SecretName)
			assert.Equal(t, tt.expected.DefaultValue, result.DefaultValue)
			assert.Equal(t, tt.expected.HasDefault, result.HasDefault)
			assert.Equal(t, tt.expected.IsSecretRef, result.IsSecretRef)
		})
	}
}

func TestParseSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []SecretRef
	}{
		{
			name:  "multiple references",
			input: []string{"sm://KEY1||default1", "sm://KEY2||default2"},
			expected: []SecretRef{
				{
					SecretName:   "KEY1",
					DefaultValue: "default1",
					HasDefault:   true,
					IsSecretRef:  true,
				},
				{
					SecretName:   "KEY2",
					DefaultValue: "default2",
					HasDefault:   true,
					IsSecretRef:  true,
				},
			},
		},
		{
			name:  "mixed references and plain values",
			input: []string{"sm://KEY1", "plain_value", "sm://KEY2||default"},
			expected: []SecretRef{
				{
					SecretName:   "KEY1",
					DefaultValue: "",
					HasDefault:   false,
					IsSecretRef:  true,
				},
				{
					SecretName:   "",
					DefaultValue: "plain_value",
					HasDefault:   true,
					IsSecretRef:  false,
				},
				{
					SecretName:   "KEY2",
					DefaultValue: "default",
					HasDefault:   true,
					IsSecretRef:  true,
				},
			},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []SecretRef{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSlice(tt.input)
			assert.Equal(t, len(tt.expected), len(result))
			for i := range tt.expected {
				assert.Equal(t, tt.expected[i].SecretName, result[i].SecretName)
				assert.Equal(t, tt.expected[i].DefaultValue, result[i].DefaultValue)
				assert.Equal(t, tt.expected[i].HasDefault, result[i].HasDefault)
				assert.Equal(t, tt.expected[i].IsSecretRef, result[i].IsSecretRef)
			}
		})
	}
}

func TestIsSecretReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "is secret reference",
			input:    "sm://API_KEY",
			expected: true,
		},
		{
			name:     "is secret reference with default",
			input:    "sm://API_KEY||default",
			expected: true,
		},
		{
			name:     "not a secret reference",
			input:    "plain_value",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSecretReference(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
