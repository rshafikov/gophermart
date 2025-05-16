package security

import (
	"regexp"
)

var LoginRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,18}[a-zA-Z0-9]$`)

func IsLoginValid(login string) bool {
	if len(login) < 3 || len(login) > 20 {
		return false
	}
	return LoginRegex.MatchString(login)
}
