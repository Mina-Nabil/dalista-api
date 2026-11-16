package listTests

import (
	"testing"

	"dalista/dalista-be/api/resources"
	listModels "dalista/dalista-be/app/repository/models/listModels"

	"github.com/google/uuid"
)

func strptr(s string) *string { return &s }

func TestItemResourceMiscAndAttributes(t *testing.T) {
	// No category → MISC bucket; typed attributes + extra map.
	misc := listModels.Item{
		ID: uuid.New(), ListID: uuid.New(), OwnerID: uuid.New(),
		Name: "Random",
		Attributes: &listModels.ItemAttributes{
			URL:   strptr("https://shop.example.com"),
			Phone: strptr("+2010"),
			Extra: `{"size":"1L","organic":true}`,
		},
	}
	r := resources.ItemFromModel(&misc)
	if !r.InMisc {
		t.Fatal("expected InMisc=true for a nil-category item")
	}
	if r.CategoryID != nil {
		t.Fatal("expected nil CategoryID for a MISC item")
	}
	if r.Attributes.URL == nil || *r.Attributes.URL != "https://shop.example.com" {
		t.Fatalf("url not mapped: %#v", r.Attributes.URL)
	}
	if r.Attributes.Phone == nil || *r.Attributes.Phone != "+2010" {
		t.Fatalf("phone not mapped: %#v", r.Attributes.Phone)
	}
	if r.Attributes.Extra["size"] != "1L" || r.Attributes.Extra["organic"] != true {
		t.Fatalf("extra not parsed: %#v", r.Attributes.Extra)
	}

	// No attributes row → nulls + empty extra, never nil map.
	none := listModels.Item{ID: uuid.New(), ListID: uuid.New(), OwnerID: uuid.New(), Name: "Bare"}
	rn := resources.ItemFromModel(&none)
	if rn.Attributes.URL != nil || rn.Attributes.Extra == nil || len(rn.Attributes.Extra) != 0 {
		t.Fatalf("expected empty attributes, got %#v", rn.Attributes)
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

func TestItemImageResource(t *testing.T) {
	item := listModels.Item{
		ID: uuid.New(), ListID: uuid.New(), OwnerID: uuid.New(), Name: "Photo item",
		Images: []listModels.ItemImage{
			{ID: uuid.New(), URL: "http://localhost:8080/media/items/x/a.png", Position: 0, SizeBytes: 123},
			{ID: uuid.New(), URL: "http://localhost:8080/media/items/x/b.png", Position: 1, SizeBytes: 456},
		},
	}
	r := resources.ItemFromModel(&item)
	if len(r.Images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(r.Images))
	}
	if r.Images[0].Position != 0 || r.Images[0].SizeBytes != 123 || r.Images[0].URL == "" {
		t.Fatalf("image[0] mapped wrong: %+v", r.Images[0])
	}
	if r.Images[1].Position != 1 {
		t.Fatalf("image[1] position wrong: %+v", r.Images[1])
	}

	// No images → non-nil empty slice (so JSON renders []).
	bare := resources.ItemFromModel(&listModels.Item{ID: uuid.New(), ListID: uuid.New(), OwnerID: uuid.New(), Name: "Bare"})
	if bare.Images == nil || len(bare.Images) != 0 {
		t.Fatalf("expected empty images slice, got %#v", bare.Images)
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
