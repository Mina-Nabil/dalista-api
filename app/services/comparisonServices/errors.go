package comparisonServices

import "errors"

var (
	// ErrNotFound: the comparison session (or share token) doesn't exist.
	ErrNotFound = errors.New("not found")
	// ErrForbidden: the caller isn't the session's owner (for owner-only actions).
	ErrForbidden = errors.New("forbidden")
	// ErrValidation: invalid input — fewer than two items, or an item the
	// creator can't access.
	ErrValidation = errors.New("validation error")
)
