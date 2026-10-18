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
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	GroupedListID *string `json:"grouped_list_id"`
	CategoryID    *string `json:"category_id"`
	IsPublic      bool    `json:"is_public"`
	IsPeriodic    bool    `json:"is_periodic"`
}

func ListFromModel(l *listModels.List) ListResource {
	return ListResource{
		ID:            l.ID.String(),
		Name:          l.Name,
		GroupedListID: uuidPtrToString(l.GroupedListID),
		CategoryID:    uuidPtrToString(l.CategoryID),
		IsPublic:      l.IsPublic,
		IsPeriodic:    l.IsPeriodic,
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
	Attributes map[string]interface{} `json:"attributes"`
}

func ItemFromModel(i *listModels.Item) ItemResource {
	attrs := map[string]interface{}{}
	if i.Attributes != "" {
		_ = json.Unmarshal([]byte(i.Attributes), &attrs)
	}
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
		Attributes: attrs,
	}
}

func ItemsFromModels(is []listModels.Item) []ItemResource {
	out := make([]ItemResource, 0, len(is))
	for i := range is {
		out = append(out, ItemFromModel(&is[i]))
	}
	return out
}
