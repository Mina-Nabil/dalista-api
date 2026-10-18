package sharingServices

import "errors"

var (
	// ErrNotFound: the list/share/invite doesn't exist or the caller isn't its owner.
	ErrNotFound = errors.New("not found")
	// ErrForbidden: the caller can see the list but lacks the role for this action.
	ErrForbidden = errors.New("forbidden")
	// ErrCapReached: the list is at its tier's collaborator cap.
	ErrCapReached = errors.New("per-list user cap reached")
	// ErrSelfShare: attempting to share a list with its own owner.
	ErrSelfShare = errors.New("cannot share a list with its owner")
	// ErrValidation: invalid input (e.g. bad role).
	ErrValidation = errors.New("validation error")
)
