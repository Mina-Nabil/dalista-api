package listTests

import (
	"testing"

	"dalista/dalista-be/api/resources"
	listModels "dalista/dalista-be/app/repository/models/listModels"

	"github.com/google/uuid"
)

func TestItemResourceMiscAndAttributes(t *testing.T) {
	// No category → MISC bucket; attributes JSON string → map.
	misc := listModels.Item{
		ID: uuid.New(), ListID: uuid.New(), OwnerID: uuid.New(),
		Name: "Random", Attributes: `{"size":"1L","organic":true}`,
	}
	r := resources.ItemFromModel(&misc)
	if !r.InMisc {
		t.Fatal("expected InMisc=true for a nil-category item")
	}
	if r.CategoryID != nil {
		t.Fatal("expected nil CategoryID for a MISC item")
	}
	if r.Attributes["size"] != "1L" || r.Attributes["organic"] != true {
		t.Fatalf("attributes not parsed: %#v", r.Attributes)
	}

	// With a category → not MISC.
	cat := uuid.New()
	withCat := listModels.Item{ID: uuid.New(), ListID: uuid.New(), OwnerID: uuid.New(), Name: "Milk", CategoryID: &cat}
	r2 := resources.ItemFromModel(&withCat)
	if r2.InMisc {
		t.Fatal("expected InMisc=false when a category is set")
	}
	if r2.CategoryID == nil || *r2.CategoryID != cat.String() {
		t.Fatalf("CategoryID = %v, want %s", r2.CategoryID, cat)
	}
}

func TestCategoryResourceIsOwned(t *testing.T) {
	owner := uuid.New()
	owned := resources.CategoryFromModel(&listModels.Category{ID: uuid.New(), Name: "Mine", OwnerID: &owner})
	if !owned.IsOwned || owned.IsSystem {
		t.Fatalf("owned category flags wrong: %+v", owned)
	}
	system := resources.CategoryFromModel(&listModels.Category{ID: uuid.New(), Name: "Groceries", IsSystem: true})
	if system.IsOwned {
		t.Fatal("system category should not be IsOwned")
	}
}
