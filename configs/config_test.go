// Package configs defines runtime configuration of the application
package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envVal   string
		exists   bool
		fallback bool
		want     bool
	}{
		{
			name:     "Valid True",
			envKey:   "TEST_BOOL_TRUE",
			envVal:   "true",
			exists:   true,
			fallback: false,
			want:     true,
		},
		{
			name:     "Valid False",
			envKey:   "TEST_BOOL_FALSE",
			envVal:   "false",
			exists:   true,
			fallback: true,
			want:     false,
		},
		{
			name:     "Invalid Value Uses Fallback",
			envKey:   "TEST_BOOL_INVALID",
			envVal:   "not-a-bool",
			exists:   true,
			fallback: true,
			want:     true,
		},
		{
			name:     "Missing Var Uses Fallback",
			envKey:   "TEST_BOOL_MISSING",
			exists:   false,
			fallback: false,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exists {
				t.Setenv(tt.envKey, tt.envVal)
			} else {
				// Ensure it's not in the env from a previous manual run
				_ = os.Unsetenv(tt.envKey)
			}

			got := EnvBool(tt.envKey, tt.fallback)
			assert.Equal(t, tt.want, got, "Case: %s", tt.name)
		})
	}
}

func TestEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envVal   string
		exists   bool
		fallback int
		want     int
	}{
		{
			name:     "Valid Positive",
			envKey:   "TEST_INT_POSITIVE",
			envVal:   "8080",
			exists:   true,
			fallback: 9090,
			want:     8080,
		},
		{
			name:     "Valid Negative",
			envKey:   "TEST_INT_NEGATIVE",
			envVal:   "-1",
			exists:   true,
			fallback: 9090,
			want:     -1,
		},
		{
			name:     "Valid Zero",
			envKey:   "TEST_INT_ZERO",
			envVal:   "0",
			exists:   true,
			fallback: 9090,
			want:     0,
		},
		{
			name:     "Invalid Value Uses Fallback",
			envKey:   "TEST_INT_INVALID",
			envVal:   "not-a-number",
			exists:   true,
			fallback: 42,
			want:     42,
		},
		{
			name:     "Missing Var Uses Fallback",
			envKey:   "TEST_INT_MISSING",
			exists:   false,
			fallback: 99,
			want:     99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exists {
				t.Setenv(tt.envKey, tt.envVal)
			} else {
				_ = os.Unsetenv(tt.envKey)
			}

			got := EnvInt(tt.envKey, tt.fallback)
			assert.Equal(t, tt.want, got, "Case: %s", tt.name)
		})
	}
}

func TestEnvString(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envVal   string
		exists   bool
		fallback string
		want     string
	}{
		{
			name:     "Valid Value",
			envKey:   "TEST_STRING_VALID",
			envVal:   "hello-world",
			exists:   true,
			fallback: "fallback",
			want:     "hello-world",
		},
		{
			name:     "Empty Value",
			envKey:   "TEST_STRING_EMPTY",
			envVal:   "",
			exists:   true,
			fallback: "fallback",
			want:     "",
		},
		{
			name:     "Missing Var Uses Fallback",
			envKey:   "TEST_STRING_MISSING",
			exists:   false,
			fallback: "fallback-value",
			want:     "fallback-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exists {
				t.Setenv(tt.envKey, tt.envVal)
			} else {
				_ = os.Unsetenv(tt.envKey)
			}

			got := EnvString(tt.envKey, tt.fallback)
			assert.Equal(t, tt.want, got, "Case: %s", tt.name)
		})
	}
}
