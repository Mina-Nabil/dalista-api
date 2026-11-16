package notificationServices

import "errors"

var (
	// ErrValidation: invalid input (bad platform, empty token, no recipients).
	ErrValidation = errors.New("validation error")
)
