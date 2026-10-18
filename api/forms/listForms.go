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

type CreateItemForm struct {
	Name       string                 `json:"name" binding:"required,max=255" example:"Milk"`
	CategoryID *string                `json:"category_id" binding:"omitempty,uuid"`
	Brand      *string                `json:"brand"`
	Price      *float64               `json:"price" example:"3.99"`
	Notes      *string                `json:"notes"`
	Attributes map[string]interface{} `json:"attributes"`
}

// UpdateItemForm is a partial update. ClearCategory moves the item to MISC.
type UpdateItemForm struct {
	Name          *string                `json:"name" binding:"omitempty,max=255"`
	CategoryID    *string                `json:"category_id" binding:"omitempty,uuid"`
	ClearCategory bool                   `json:"clear_category"`
	Brand         *string                `json:"brand"`
	Price         *float64               `json:"price"`
	Notes         *string                `json:"notes"`
	Attributes    map[string]interface{} `json:"attributes"`
}
