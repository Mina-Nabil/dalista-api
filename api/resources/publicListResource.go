package resources

import (
	publicListServices "dalista/dalista-be/app/services/publicListServices"
)

// DiscoveryItemResource is a public list in the discovery feed, plus whether the
// caller has starred it. Embeds ListResource (which already carries star_count).
type DiscoveryItemResource struct {
	ListResource
	StarredByMe bool `json:"starred_by_me"`
}

// DiscoveryFeedResource is a paginated page of the discovery feed.
type DiscoveryFeedResource struct {
	Items  []DiscoveryItemResource `json:"items"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
	Total  int64                   `json:"total"`
}

// StarResource is the result of a star/unstar: the caller's current state and
// the list's star count.
type StarResource struct {
	ListID    string `json:"list_id"`
	Starred   bool   `json:"starred"`
	StarCount int    `json:"star_count"`
}

// DiscoveryFeedFromModels maps a discovery page to its resource.
func DiscoveryFeedFromModels(items []publicListServices.DiscoveryItem, limit, offset int, total int64) DiscoveryFeedResource {
	out := DiscoveryFeedResource{
		Items:  make([]DiscoveryItemResource, 0, len(items)),
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}
	for i := range items {
		l := items[i].List
		out.Items = append(out.Items, DiscoveryItemResource{
			ListResource: ListFromModel(&l),
			StarredByMe:  items[i].StarredByMe,
		})
	}
	return out
}

// StarResourceFrom maps a star state to its resource.
func StarResourceFrom(listID string, s publicListServices.StarState) StarResource {
	return StarResource{ListID: listID, Starred: s.Starred, StarCount: s.StarCount}
}
