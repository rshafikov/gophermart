package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsLoginValid(t *testing.T) {
	tests := []struct {
		login string
		valid bool
	}{
		{"testuser", true},
		{"user_123", true},
		{"a-b_c123", true},
		{"ab", false},
		{"user@123", false},
		{"_user", false},
		{"user_", false},
		{"toolongusername123456", false},
	}

	for _, test := range tests {
		t.Run("test login:"+test.login, func(t *testing.T) {
			valid := IsLoginValid(test.login)
			assert.Equal(t, test.valid, valid)
		})
	}
}
