package publicListTests

import (
	"testing"

	"dalista/dalista-be/api/resources"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	publicListServices "dalista/dalista-be/app/services/publicListServices"

	"github.com/google/uuid"
)

func TestClampLimit(t *testing.T) {
	cases := map[int]int{
		0:    publicListServices.DefaultLimit,
		-5:   publicListServices.DefaultLimit,
		10:   10,
		100:  100,
		1000: publicListServices.MaxLimit,
	}
	for in, want := range cases {
		if got := publicListServices.ClampLimit(in); got != want {
			t.Fatalf("ClampLimit(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestDiscoveryFeedResourceMapping(t *testing.T) {
	l1 := listModels.List{ID: uuid.New(), Name: "Top", IsPublic: true, StarCount: 9}
	l2 := listModels.List{ID: uuid.New(), Name: "Next", IsPublic: true, StarCount: 2}
	items := []publicListServices.DiscoveryItem{
		{List: l1, StarredByMe: true},
		{List: l2, StarredByMe: false},
	}

	r := resources.DiscoveryFeedFromModels(items, 20, 40, 57)
	if r.Limit != 20 || r.Offset != 40 || r.Total != 57 {
		t.Fatalf("pagination fields wrong: %+v", r)
	}
	if len(r.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(r.Items))
	}
	if r.Items[0].Name != "Top" || r.Items[0].StarCount != 9 || !r.Items[0].StarredByMe {
		t.Fatalf("item 0 mapping wrong: %+v", r.Items[0])
	}
	if !r.Items[0].IsPublic {
		t.Fatalf("item 0 should be public")
	}
	if r.Items[1].StarredByMe {
		t.Fatalf("item 1 should not be starred by me")
	}
}

func TestStarResourceMapping(t *testing.T) {
	id := uuid.New().String()
	r := resources.StarResourceFrom(id, publicListServices.StarState{Starred: true, StarCount: 3})
	if r.ListID != id || !r.Starred || r.StarCount != 3 {
		t.Fatalf("star resource mapping wrong: %+v", r)
	}
	r2 := resources.StarResourceFrom(id, publicListServices.StarState{Starred: false, StarCount: 2})
	if r2.Starred || r2.StarCount != 2 {
		t.Fatalf("unstar resource mapping wrong: %+v", r2)
	}
}
