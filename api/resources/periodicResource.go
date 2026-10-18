package resources

import (
	"encoding/json"
	"time"

	periodicModels "dalista/dalista-be/app/repository/models/periodicModels"
)

// VariantItemResource is a snapshot item within a list variant.
type VariantItemResource struct {
	ID           string                 `json:"id"`
	SourceItemID *string                `json:"source_item_id"`
	Name         string                 `json:"name"`
	Brand        *string                `json:"brand"`
	Price        *float64               `json:"price"`
	CategoryID   *string                `json:"category_id"`
	Notes        *string                `json:"notes"`
	Attributes   map[string]interface{} `json:"attributes"`
}

// ListVariantResource is a dated "today's list" snapshot.
type ListVariantResource struct {
	ID             string                `json:"id"`
	PeriodicListID string                `json:"periodic_list_id"`
	Title          string                `json:"title"`
	GeneratedBy    *string               `json:"generated_by"`
	GeneratedAt    string                `json:"generated_at"`
	ItemCount      int                   `json:"item_count"`
	Items          []VariantItemResource `json:"items,omitempty"`
}

func variantItemFromModel(i *periodicModels.ListVariantItem) VariantItemResource {
	attrs := map[string]interface{}{}
	if i.Attributes != "" {
		_ = json.Unmarshal([]byte(i.Attributes), &attrs)
	}
	return VariantItemResource{
		ID:           i.ID.String(),
		SourceItemID: idPtrString(i.SourceItemID),
		Name:         i.Name,
		Brand:        i.Brand,
		Price:        i.Price,
		CategoryID:   idPtrString(i.CategoryID),
		Notes:        i.Notes,
		Attributes:   attrs,
	}
}

// ListVariantFromModel maps a variant including its snapshot items.
func ListVariantFromModel(v *periodicModels.ListVariant) ListVariantResource {
	res := variantSummary(v)
	res.Items = make([]VariantItemResource, 0, len(v.Items))
	for i := range v.Items {
		res.Items = append(res.Items, variantItemFromModel(&v.Items[i]))
	}
	return res
}

// ListVariantsSummary maps variants to summaries (count only, no item list).
func ListVariantsSummary(vs []periodicModels.ListVariant) []ListVariantResource {
	out := make([]ListVariantResource, 0, len(vs))
	for i := range vs {
		out = append(out, variantSummary(&vs[i]))
	}
	return out
}

func variantSummary(v *periodicModels.ListVariant) ListVariantResource {
	return ListVariantResource{
		ID:             v.ID.String(),
		PeriodicListID: v.PeriodicListID.String(),
		Title:          v.Title,
		GeneratedBy:    idPtrString(v.GeneratedBy),
		GeneratedAt:    v.GeneratedAt.UTC().Format(time.RFC3339),
		ItemCount:      len(v.Items),
	}
}
