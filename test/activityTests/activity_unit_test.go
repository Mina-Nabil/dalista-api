package activityTests

import (
	"testing"

	"dalista/dalista-be/api/resources"
	activityModels "dalista/dalista-be/app/repository/models/activityModels"

	"github.com/google/uuid"
)

func TestActivityEventResourceParsesPayload(t *testing.T) {
	actor := uuid.New()
	target := uuid.New()
	ev := activityModels.ActivityEvent{
		ID: uuid.New(), ListID: uuid.New(), ActorID: &actor,
		Action: "item_created", TargetType: "item", TargetID: &target,
		Payload: `{"name":"Milk"}`,
	}
	r := resources.ActivityEventFromModel(&ev)
	if r.Action != "item_created" || r.Payload["name"] != "Milk" {
		t.Fatalf("event mapping wrong: %+v", r)
	}
	if r.ActorID == nil || *r.ActorID != actor.String() || r.TargetID == nil || *r.TargetID != target.String() {
		t.Fatalf("actor/target id mapping wrong: %+v", r)
	}
}

func TestCommentResourceSystemHasNilAuthor(t *testing.T) {
	sys := activityModels.Comment{ID: uuid.New(), ItemID: uuid.New(), Body: "Marked as need to buy", IsSystem: true}
	r := resources.CommentFromModel(&sys)
	if !r.IsSystem || r.AuthorID != nil {
		t.Fatalf("system comment should have nil author: %+v", r)
	}
	author := uuid.New()
	usr := activityModels.Comment{ID: uuid.New(), ItemID: uuid.New(), AuthorID: &author, Body: "hi"}
	r2 := resources.CommentFromModel(&usr)
	if r2.IsSystem || r2.AuthorID == nil || *r2.AuthorID != author.String() {
		t.Fatalf("user comment mapping wrong: %+v", r2)
	}
}
