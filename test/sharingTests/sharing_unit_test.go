package sharingTests

import (
	"testing"

	"dalista/dalista-be/api/resources"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	sharingModels "dalista/dalista-be/app/repository/models/sharingModels"
	sharingServices "dalista/dalista-be/app/services/sharingServices"

	"github.com/google/uuid"
)

func listModelWith(name string) listModels.List {
	return listModels.List{ID: uuid.New(), OwnerID: uuid.New(), Name: name}
}

func TestShareResultResource_Share(t *testing.T) {
	uid := uuid.New()
	lid := uuid.New()
	res := &sharingServices.ShareResult{
		Kind:  sharingServices.KindShare,
		Share: &sharingModels.ListShare{ListID: lid, UserID: uid, Role: sharingModels.RoleEditor},
	}
	r := resources.ShareResultFromModel(res)
	if r.Kind != "share" || r.UserID == nil || *r.UserID != uid.String() || r.Phone != nil || r.Role != "editor" {
		t.Fatalf("share mapping wrong: %+v", r)
	}
	if r.ListID != lid.String() {
		t.Fatalf("list id = %s, want %s", r.ListID, lid)
	}
}

func TestShareResultResource_Invite(t *testing.T) {
	res := &sharingServices.ShareResult{
		Kind:   sharingServices.KindInvite,
		Invite: &sharingModels.PendingInvite{ListID: uuid.New(), Phone: "+201001234567", Role: sharingModels.RoleViewer},
	}
	r := resources.ShareResultFromModel(res)
	if r.Kind != "invite" || r.Phone == nil || *r.Phone != "+201001234567" || r.UserID != nil || r.Role != "viewer" {
		t.Fatalf("invite mapping wrong: %+v", r)
	}
}

func TestSharedListsResource(t *testing.T) {
	rows := []sharingServices.SharedList{
		{List: listModelWith("Groceries"), Role: sharingModels.RoleEditor},
	}
	out := resources.SharedListsFromModels(rows)
	if len(out) != 1 || out[0].Name != "Groceries" || out[0].Role != "editor" {
		t.Fatalf("shared list mapping wrong: %+v", out)
	}
}
