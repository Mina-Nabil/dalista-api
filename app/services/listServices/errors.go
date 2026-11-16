package listServices

import "errors"

var (
	// ErrNotFound means the record doesn't exist or the caller has no access.
	// No-access is reported as not-found to avoid leaking existence.
	ErrNotFound = errors.New("not found")
	// ErrValidation is a domain validation failure (e.g. referencing a grouped
	// list or category the caller can't use).
	ErrValidation = errors.New("validation error")
	// ErrForbidden means the caller can see the resource but lacks the role to
	// perform the action (e.g. a viewer attempting a mutation).
	ErrForbidden = errors.New("forbidden")
	// ErrUnsupportedMediaType means an uploaded image's content type isn't allowed.
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	// ErrFileTooLarge means an uploaded file exceeds the per-image size limit.
	ErrFileTooLarge = errors.New("file too large")
	// ErrTooManyImages means the item already has the maximum number of images.
	ErrTooManyImages = errors.New("too many images")
)
