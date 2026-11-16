package comparisonTests

import (
	"testing"
	"time"

	"dalista/dalista-be/api/resources"
	comparisonModels "dalista/dalista-be/app/repository/models/comparisonModels"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	comparisonServices "dalista/dalista-be/app/services/comparisonServices"

	"github.com/google/uuid"
)

func ptr(s string) *string { return &s }

func TestComparisonFromViewMapsItemsAndWithheld(t *testing.T) {
	owner := uuid.New()
	token := "abc123"
	session := &comparisonModels.ComparisonSession{
		ID:         uuid.New(),
		OwnerID:    owner,
		ShareToken: ptr(token),
		CreatedAt:  time.Now(),
	}
	price := 12.5
	view := &comparisonServices.View{
		Session: session,
		Items: []listModels.Item{
			{ID: uuid.New(), ListID: uuid.New(), Name: "Olive oil", Brand: ptr("Acme"), Price: &price},
			{ID: uuid.New(), ListID: uuid.New(), Name: "Olive oil XL"},
		},
		Withheld: 3,
	}

	r := resources.ComparisonFromView(view)
	if r.ID != session.ID.String() || r.OwnerID != owner.String() {
		t.Fatalf("id/owner mapping wrong: %+v", r)
	}
	if r.ShareToken == nil || *r.ShareToken != token {
		t.Fatalf("share token mapping wrong: %+v", r.ShareToken)
	}
	if r.ItemCount != 2 || len(r.Items) != 2 {
		t.Fatalf("item count/items wrong: count=%d items=%d", r.ItemCount, len(r.Items))
	}
	if r.Withheld != 3 {
		t.Fatalf("withheld should be 3, got %d", r.Withheld)
	}
	if r.Items[0].Name != "Olive oil" || r.Items[0].Price == nil || *r.Items[0].Price != 12.5 {
		t.Fatalf("item 0 mapping wrong: %+v", r.Items[0])
	}
}

func TestComparisonFromViewNilTokenEmptyItems(t *testing.T) {
	session := &comparisonModels.ComparisonSession{ID: uuid.New(), OwnerID: uuid.New(), CreatedAt: time.Now()}
	r := resources.ComparisonFromView(&comparisonServices.View{Session: session, Items: nil, Withheld: 0})
	if r.ShareToken != nil {
		t.Fatalf("expected nil share token, got %v", r.ShareToken)
	}
	if r.ItemCount != 0 || len(r.Items) != 0 {
		t.Fatalf("expected no items, got count=%d len=%d", r.ItemCount, len(r.Items))
	}
}
