package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Valid password",
			password:    "securepass123",
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    "",
			expectError: false,
		},
		{
			name:        "Short password",
			password:    "short",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if tt.expectError {
				assert.Error(t, err, "Expected error for password: %s", tt.password)
				return
			}
			assert.NoError(t, err, "Unexpected error for password: %s", tt.password)
			assert.NotEmpty(t, hash, "Hash should not be empty")

			if tt.password != "" {
				assert.True(t, CheckPasswordHash(tt.password, hash), "Hash should match the password")
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "securepass123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "Correct password",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "Incorrect password",
			password: "wrongpass",
			hash:     hash,
			expected: false,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "Invalid hash",
			password: password,
			hash:     "invalid-hash",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPasswordHash(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result, "CheckPasswordHash failed for: %s", tt.name)
		})
	}
}

func TestIsPasswordValid(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "Valid password (8 chars)",
			password: "password",
			expected: true,
		},
		{
			name:     "Valid password (>8 chars)",
			password: "securepassword123",
			expected: true,
		},
		{
			name:     "Invalid password (7 chars)",
			password: "passwor",
			expected: false,
		},
		{
			name:     "Empty password",
			password: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPasswordValid(tt.password)
			assert.Equal(t, tt.expected, result, "IsPasswordValid failed for: %s", tt.name)
		})
	}
}
