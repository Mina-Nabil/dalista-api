package periodicTests

import (
	"testing"

	"dalista/dalista-be/api/resources"
	listModels "dalista/dalista-be/app/repository/models/listModels"
	periodicModels "dalista/dalista-be/app/repository/models/periodicModels"

	"github.com/google/uuid"
)

func strPtr(s string) *string   { return &s }
func f64Ptr(f float64) *float64 { return &f }

func TestItemResourceCarriesNeedToBuy(t *testing.T) {
	item := &listModels.Item{ID: uuid.New(), ListID: uuid.New(), Name: "Milk", NeedToBuy: true}
	r := resources.ItemFromModel(item)
	if !r.NeedToBuy {
		t.Fatalf("expected need_to_buy true, got %+v", r)
	}
	item.NeedToBuy = false
	if resources.ItemFromModel(item).NeedToBuy {
		t.Fatalf("expected need_to_buy false")
	}
}

func TestListVariantFromModelMapsItems(t *testing.T) {
	actor := uuid.New()
	src := uuid.New()
	cat := uuid.New()
	v := &periodicModels.ListVariant{
		ID:             uuid.New(),
		PeriodicListID: uuid.New(),
		Title:          "Groceries — 2026-10-21",
		GeneratedBy:    &actor,
		Items: []periodicModels.ListVariantItem{
			{
				ID:           uuid.New(),
				SourceItemID: &src,
				Name:         "Milk",
				Brand:        strPtr("Juhayna"),
				Price:        f64Ptr(42.5),
				CategoryID:   &cat,
				Notes:        strPtr("2 cartons"),
				Attributes:   `{"size":"1L"}`,
			},
		},
	}

	r := resources.ListVariantFromModel(v)
	if r.Title != "Groceries — 2026-10-21" {
		t.Fatalf("title mapping wrong: %q", r.Title)
	}
	if r.GeneratedBy == nil || *r.GeneratedBy != actor.String() {
		t.Fatalf("generated_by mapping wrong: %+v", r.GeneratedBy)
	}
	if r.ItemCount != 1 || len(r.Items) != 1 {
		t.Fatalf("item count/items wrong: count=%d items=%d", r.ItemCount, len(r.Items))
	}
	it := r.Items[0]
	if it.Name != "Milk" || it.SourceItemID == nil || *it.SourceItemID != src.String() {
		t.Fatalf("item/source mapping wrong: %+v", it)
	}
	if it.Brand == nil || *it.Brand != "Juhayna" || it.Price == nil || *it.Price != 42.5 {
		t.Fatalf("brand/price mapping wrong: %+v", it)
	}
	if it.CategoryID == nil || *it.CategoryID != cat.String() {
		t.Fatalf("category mapping wrong: %+v", it)
	}
	if it.Attributes["size"] != "1L" {
		t.Fatalf("attributes mapping wrong: %+v", it.Attributes)
	}
}

func TestListVariantsSummaryOmitsItems(t *testing.T) {
	vs := []periodicModels.ListVariant{
		{
			ID:             uuid.New(),
			PeriodicListID: uuid.New(),
			Title:          "Run A",
			Items:          []periodicModels.ListVariantItem{{ID: uuid.New(), Name: "A"}, {ID: uuid.New(), Name: "B"}},
		},
	}
	out := resources.ListVariantsSummary(vs)
	if len(out) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(out))
	}
	if out[0].ItemCount != 2 {
		t.Fatalf("expected item_count 2, got %d", out[0].ItemCount)
	}
	if len(out[0].Items) != 0 {
		t.Fatalf("summary should not include items, got %d", len(out[0].Items))
	}
}
