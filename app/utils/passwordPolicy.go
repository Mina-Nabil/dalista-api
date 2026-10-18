package utils

import "unicode/utf8"

// PasswordMeetsPolicy reports whether s satisfies Dalista's password rules:
// at least 8 characters, with at least one letter and one digit. (Consumer-app
// policy — lighter than Lumina's 12-char HIPAA rule.)
func PasswordMeetsPolicy(s string) bool {
	if utf8.RuneCountInString(s) < 8 {
		return false
	}
	var letter, digit bool
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			letter = true
		case r >= '0' && r <= '9':
			digit = true
		}
	}
	return letter && digit
}
