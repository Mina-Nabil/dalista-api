package resources

import (
	sharingServices "dalista/dalista-be/app/services/sharingServices"
)

// ShareResultResource is the response to sharing a list: whether it created an
// immediate share (registered user) or a pending invite (unregistered phone).
type ShareResultResource struct {
	Kind   string  `json:"kind"` // "share" or "invite"
	ListID string  `json:"list_id"`
	UserID *string `json:"user_id"`
	Phone  *string `json:"phone"`
	Role   string  `json:"role"`
}

// ShareResultFromModel maps a sharing service result to its resource.
func ShareResultFromModel(r *sharingServices.ShareResult) ShareResultResource {
	out := ShareResultResource{Kind: string(r.Kind)}
	switch r.Kind {
	case sharingServices.KindShare:
		uid := r.Share.UserID.String()
		out.ListID = r.Share.ListID.String()
		out.UserID = &uid
		out.Role = string(r.Share.Role)
	case sharingServices.KindInvite:
		phone := r.Invite.Phone
		out.ListID = r.Invite.ListID.String()
		out.Phone = &phone
		out.Role = string(r.Invite.Role)
	}
	return out
}

// SharedListResource is a list shared with the caller, plus the caller's role.
type SharedListResource struct {
	ListResource
	Role string `json:"role"`
}

// SharedListsFromModels maps shared-with-me rows to resources.
func SharedListsFromModels(rows []sharingServices.SharedList) []SharedListResource {
	out := make([]SharedListResource, 0, len(rows))
	for i := range rows {
		l := rows[i].List
		out = append(out, SharedListResource{
			ListResource: ListFromModel(&l),
			Role:         string(rows[i].Role),
		})
	}
	return out
}
