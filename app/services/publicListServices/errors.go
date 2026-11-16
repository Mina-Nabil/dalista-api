package publicListServices

import "errors"

var (
	// ErrNotFound: the list doesn't exist.
	ErrNotFound = errors.New("not found")
	// ErrNotPublic: the list exists but isn't public, so it can't be starred or discovered.
	ErrNotPublic = errors.New("list is not public")
)
