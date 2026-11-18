package forms

// NeedToBuyForm toggles an item's need-to-buy state. Pointer so `false` is explicit.
// Quantity is optional: when marking need-to-buy it seeds the count (defaults to
// 1 when omitted); it is ignored when unmarking (the count resets to 1).
type NeedToBuyForm struct {
	NeedToBuy *bool `json:"need_to_buy" binding:"required" example:"true"`
	Quantity  *int  `json:"quantity" binding:"omitempty,min=1" example:"1"`
}

// QuantityForm sets an item's need-to-buy count on a periodic list. Pointer +
// required so a missing body is rejected rather than defaulting to 0.
type QuantityForm struct {
	Quantity *int `json:"quantity" binding:"required,min=1" example:"2"`
}

// GenerateTodayForm confirms a "today's list": any item ids listed here are
// excluded (marked "enough"); the rest of the need-to-buy set forms the variant.
type GenerateTodayForm struct {
	ExcludeItemIDs []string `json:"exclude_item_ids" binding:"omitempty,dive,uuid"`
}
