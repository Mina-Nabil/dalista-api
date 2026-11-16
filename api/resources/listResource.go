package resources

import (
	"encoding/json"

	listModels "dalista/dalista-be/app/repository/models/listModels"

	"github.com/google/uuid"
)

// uuidPtrToString maps a *uuid.UUID to a *string (nil stays nil) for JSON.
func uuidPtrToString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// --- Category ---

type CategoryResource struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsSystem bool   `json:"is_system"`
	IsOwned  bool   `json:"is_owned"`
}

func CategoryFromModel(c *listModels.Category) CategoryResource {
	return CategoryResource{
		ID:       c.ID.String(),
		Name:     c.Name,
		IsSystem: c.IsSystem,
		IsOwned:  c.OwnerID != nil,
	}
}

func CategoriesFromModels(cs []listModels.Category) []CategoryResource {
	out := make([]CategoryResource, 0, len(cs))
	for i := range cs {
		out = append(out, CategoryFromModel(&cs[i]))
	}
	return out
}

// --- Grouped List ---

type GroupedListResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GroupedListFromModel(g *listModels.GroupedList) GroupedListResource {
	return GroupedListResource{ID: g.ID.String(), Name: g.Name}
}

func GroupedListsFromModels(gs []listModels.GroupedList) []GroupedListResource {
	out := make([]GroupedListResource, 0, len(gs))
	for i := range gs {
		out = append(out, GroupedListFromModel(&gs[i]))
	}
	return out
}

// --- List ---

type ListResource struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	GroupedListID        *string `json:"grouped_list_id"`
	CategoryID           *string `json:"category_id"`
	SourcePeriodicListID *string `json:"source_periodic_list_id"`
	IsPublic             bool    `json:"is_public"`
	IsPeriodic           bool    `json:"is_periodic"`
	StarCount            int     `json:"star_count"`
}

func ListFromModel(l *listModels.List) ListResource {
	return ListResource{
		ID:                   l.ID.String(),
		Name:                 l.Name,
		GroupedListID:        uuidPtrToString(l.GroupedListID),
		CategoryID:           uuidPtrToString(l.CategoryID),
		SourcePeriodicListID: uuidPtrToString(l.SourcePeriodicListID),
		IsPublic:             l.IsPublic,
		IsPeriodic:           l.IsPeriodic,
		StarCount:            l.StarCount,
	}
}

func ListsFromModels(ls []listModels.List) []ListResource {
	out := make([]ListResource, 0, len(ls))
	for i := range ls {
		out = append(out, ListFromModel(&ls[i]))
	}
	return out
}

// --- Item ---

type ItemResource struct {
	ID         string                 `json:"id"`
	ListID     string                 `json:"list_id"`
	CategoryID *string                `json:"category_id"`
	InMisc     bool                   `json:"in_misc"`
	Name       string                 `json:"name"`
	Brand      *string                `json:"brand"`
	Price      *float64               `json:"price"`
	Notes      *string                `json:"notes"`
	NeedToBuy  bool                   `json:"need_to_buy"`
	IsBought   bool                   `json:"is_bought"`
	Attributes ItemAttributesResource `json:"attributes"`
	Images     []ItemImageResource    `json:"images"`
}

// ItemImageResource is one photo attached to an item.
type ItemImageResource struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Position  int    `json:"position"`
	SizeBytes int64  `json:"size_bytes"`
}

func ItemImageFromModel(i *listModels.ItemImage) ItemImageResource {
	return ItemImageResource{
		ID:        i.ID.String(),
		URL:       i.URL,
		Position:  i.Position,
		SizeBytes: i.SizeBytes,
	}
}

func ItemImagesFromModels(is []listModels.ItemImage) []ItemImageResource {
	out := make([]ItemImageResource, 0, len(is))
	for i := range is {
		out = append(out, ItemImageFromModel(&is[i]))
	}
	return out
}

// ItemAttributesResource is the typed, well-known attribute set plus an `extra`
// map for custom key/values. Always present on an item (nulls + empty extra
// when the item has no attributes row).
type ItemAttributesResource struct {
	URL       *string                `json:"url"`
	Location  *string                `json:"location"`
	Facebook  *string                `json:"facebook"`
	Instagram *string                `json:"instagram"`
	Phone     *string                `json:"phone"`
	Extra     map[string]interface{} `json:"extra"`
}

func itemAttributesFromModel(a *listModels.ItemAttributes) ItemAttributesResource {
	extra := map[string]interface{}{}
	if a == nil {
		return ItemAttributesResource{Extra: extra}
	}
	if a.Extra != "" {
		_ = json.Unmarshal([]byte(a.Extra), &extra)
	}
	return ItemAttributesResource{
		URL:       a.URL,
		Location:  a.Location,
		Facebook:  a.Facebook,
		Instagram: a.Instagram,
		Phone:     a.Phone,
		Extra:     extra,
	}
}

func ItemFromModel(i *listModels.Item) ItemResource {
	return ItemResource{
		ID:         i.ID.String(),
		ListID:     i.ListID.String(),
		CategoryID: uuidPtrToString(i.CategoryID),
		InMisc:     i.CategoryID == nil,
		Name:       i.Name,
		Brand:      i.Brand,
		Price:      i.Price,
		Notes:      i.Notes,
		NeedToBuy:  i.NeedToBuy,
		IsBought:   i.IsBought,
		Attributes: itemAttributesFromModel(i.Attributes),
		Images:     ItemImagesFromModels(i.Images),
	}
}

func ItemsFromModels(is []listModels.Item) []ItemResource {
	out := make([]ItemResource, 0, len(is))
	for i := range is {
		out = append(out, ItemFromModel(&is[i]))
	}
	return out
}
