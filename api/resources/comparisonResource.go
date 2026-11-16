package resources

import (
	"time"

	comparisonServices "dalista/dalista-be/app/services/comparisonServices"
)

// ComparisonResource is a resolved comparison: the (access-gated) items in the
// owner's selection order, plus how many were withheld from this viewer and the
// share token if the session has been shared.
type ComparisonResource struct {
	ID         string         `json:"id"`
	OwnerID    string         `json:"owner_id"`
	ShareToken *string        `json:"share_token"`
	CreatedAt  string         `json:"created_at"`
	ItemCount  int            `json:"item_count"`
	Withheld   int            `json:"withheld"`
	Items      []ItemResource `json:"items"`
}

// ComparisonFromView maps a resolved comparison view to its resource.
func ComparisonFromView(v *comparisonServices.View) ComparisonResource {
	s := v.Session
	res := ComparisonResource{
		ID:         s.ID.String(),
		OwnerID:    s.OwnerID.String(),
		ShareToken: s.ShareToken,
		CreatedAt:  s.CreatedAt.UTC().Format(time.RFC3339),
		ItemCount:  len(v.Items),
		Withheld:   v.Withheld,
		Items:      ItemsFromModels(v.Items),
	}
	return res
}
