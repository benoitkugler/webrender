package utils

import (
	"unicode"
)

// Return True if all characters in S are digits and there is at least one character in S, False otherwise.
func IsDigit(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
