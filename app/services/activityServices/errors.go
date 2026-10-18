package activityServices

import "errors"

var (
	// ErrNotFound: the comment doesn't exist or the caller can't act on it.
	ErrNotFound = errors.New("not found")
	// ErrForbidden: the caller lacks permission for this action on the comment.
	ErrForbidden = errors.New("forbidden")
)
