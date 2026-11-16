package forms

// CreateComparisonForm selects the items to compare. At least two item IDs are
// required; they must be items the caller can access (own, shared, or public
// catalog). Order is preserved; duplicates are collapsed.
type CreateComparisonForm struct {
	ItemIDs []string `json:"item_ids" binding:"required,min=2,dive,uuid"`
}
