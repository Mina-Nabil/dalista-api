package forms

// --- Category ---

type CategoryForm struct {
	Name string `json:"name" binding:"required,max=255" example:"Groceries"`
}

// --- Grouped List ---

type GroupedListForm struct {
	Name string `json:"name" binding:"required,max=255" example:"Home"`
}

// --- List ---

// CreateListForm is the body for creating a list. UUID references are strings
// and parsed/validated in the handler.
type CreateListForm struct {
	Name          string  `json:"name" binding:"required,max=255" example:"Weekly groceries"`
	GroupedListID *string `json:"grouped_list_id" binding:"omitempty,uuid"`
	CategoryID    *string `json:"category_id" binding:"omitempty,uuid"`
	IsPublic      *bool   `json:"is_public"`
	IsPeriodic    *bool   `json:"is_periodic"`
}

// UpdateListForm is a partial update. Clear flags move the list to standalone /
// clear its category (since a nil pointer means "leave unchanged").
type UpdateListForm struct {
	Name             *string `json:"name" binding:"omitempty,max=255"`
	GroupedListID    *string `json:"grouped_list_id" binding:"omitempty,uuid"`
	CategoryID       *string `json:"category_id" binding:"omitempty,uuid"`
	IsPublic         *bool   `json:"is_public"`
	IsPeriodic       *bool   `json:"is_periodic"`
	ClearGroupedList bool    `json:"clear_grouped_list"`
	ClearCategory    bool    `json:"clear_category"`
}

// --- Item ---

// BoughtForm toggles an item's generic "bought / included" checkbox.
// Pointer so an explicit `false` is distinguishable from omitted.
type BoughtForm struct {
	IsBought *bool `json:"is_bought" binding:"required"`
}

type CreateItemForm struct {
	Name       string   `json:"name" binding:"required,max=255" example:"Milk"`
	CategoryID *string  `json:"category_id" binding:"omitempty,uuid"`
	Brand      *string  `json:"brand"`
	Price      *float64 `json:"price" example:"3.99"`
	Notes      *string  `json:"notes"`
	// Optional well-known attributes. Each is independent; omit to leave unset,
	// send "" to clear. URL is format-checked when non-empty.
	URL       *string                `json:"url" example:"https://shop.example.com/item"`
	Location  *string                `json:"location"`
	Facebook  *string                `json:"facebook"`
	Instagram *string                `json:"instagram"`
	Phone     *string                `json:"phone"`
	// Extra holds arbitrary custom key/values (the XD "Add Other" row).
	Extra map[string]interface{} `json:"extra"`
}

// UpdateItemForm is a partial update. ClearCategory moves the item to MISC.
// For the attribute fields, omit to leave a column unchanged; send "" to clear
// it; send the whole `extra` object to replace the custom map.
type UpdateItemForm struct {
	Name          *string  `json:"name" binding:"omitempty,max=255"`
	CategoryID    *string  `json:"category_id" binding:"omitempty,uuid"`
	ClearCategory bool     `json:"clear_category"`
	Brand         *string  `json:"brand"`
	Price         *float64 `json:"price"`
	Notes         *string  `json:"notes"`
	URL           *string  `json:"url"`
	Location      *string  `json:"location"`
	Facebook      *string  `json:"facebook"`
	Instagram     *string  `json:"instagram"`
	Phone         *string  `json:"phone"`
	Extra         map[string]interface{} `json:"extra"`
}
