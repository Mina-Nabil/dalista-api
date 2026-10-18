package forms

// NeedToBuyForm toggles an item's need-to-buy state. Pointer so `false` is explicit.
type NeedToBuyForm struct {
	NeedToBuy *bool `json:"need_to_buy" binding:"required" example:"true"`
}

// GenerateTodayForm confirms a "today's list": any item ids listed here are
// excluded (marked "enough"); the rest of the need-to-buy set forms the variant.
type GenerateTodayForm struct {
	ExcludeItemIDs []string `json:"exclude_item_ids" binding:"omitempty,dive,uuid"`
}
