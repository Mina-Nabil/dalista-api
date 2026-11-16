package periodicServices

import "errors"

var (
	// ErrNotFound: the variant (or item) doesn't exist / isn't accessible.
	ErrNotFound = errors.New("not found")
	// ErrNotPeriodic: the operation requires a periodic list.
	ErrNotPeriodic = errors.New("list is not periodic")
	// ErrNoItems: no need-to-buy items to put in a today's list.
	ErrNoItems = errors.New("no items marked need to buy")
)
