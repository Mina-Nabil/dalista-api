package billingServices

import (
	"context"

	"dalista/dalista-be/app/repository"
	billingModels "dalista/dalista-be/app/repository/models/billingModels"
)

// SubscriptionSummary aggregates subscription counts for the dashboard's
// oversight view: by tier and by billing rail.
type SubscriptionSummary struct {
	Total      int            `json:"total"`
	ByTier     map[string]int `json:"by_tier"`
	ByProvider map[string]int `json:"by_provider"`
	ByStatus   map[string]int `json:"by_status"`
}

// OversightRow is one subscription enriched with the owner's contact fields.
type OversightRow struct {
	UserID           string  `gorm:"column:user_id" json:"user_id"`
	Email            string  `gorm:"column:email" json:"email"`
	Phone            string  `gorm:"column:phone" json:"phone"`
	BillingProvider  string  `gorm:"column:billing_provider" json:"billing_provider"`
	Tier             string  `gorm:"column:tier" json:"tier"`
	Status           string  `gorm:"column:status" json:"status"`
	ProductID        *string `gorm:"column:product_id" json:"product_id"`
	CurrentPeriodEnd *string `gorm:"column:current_period_end" json:"current_period_end"`
}

// Summary returns grouped counts across all subscriptions.
func Summary(ctx context.Context) (SubscriptionSummary, error) {
	db := repository.GetDB().WithContext(ctx)
	out := SubscriptionSummary{ByTier: map[string]int{}, ByProvider: map[string]int{}, ByStatus: map[string]int{}}

	type gcount struct {
		Key   string `gorm:"column:key"`
		Count int    `gorm:"column:count"`
	}
	group := func(col string, dst map[string]int) error {
		var rows []gcount
		if err := db.Model(&billingModels.Subscription{}).
			Select(col + " AS key, COUNT(*) AS count").Group(col).Scan(&rows).Error; err != nil {
			return err
		}
		for _, r := range rows {
			dst[r.Key] = r.Count
		}
		return nil
	}
	if err := group("tier", out.ByTier); err != nil {
		return out, err
	}
	if err := group("billing_provider", out.ByProvider); err != nil {
		return out, err
	}
	if err := group("status", out.ByStatus); err != nil {
		return out, err
	}
	for _, n := range out.ByTier {
		out.Total += n
	}
	return out, nil
}

// List returns subscriptions (optionally filtered by tier/provider) joined to
// the owner's contact fields, for the oversight table.
func List(ctx context.Context, tier, provider string) ([]OversightRow, error) {
	q := repository.GetDB().WithContext(ctx).
		Table("subscriptions s").
		Select(`s.user_id AS user_id, COALESCE(u.email,'') AS email, COALESCE(u.phone,'') AS phone,
		         s.billing_provider AS billing_provider, s.tier AS tier, s.status AS status,
		         s.product_id AS product_id, to_char(s.current_period_end, 'YYYY-MM-DD"T"HH24:MI:SSZ') AS current_period_end`).
		Joins("JOIN users u ON u.id = s.user_id").
		Order("s.updated_at DESC")
	if tier != "" {
		q = q.Where("s.tier = ?", tier)
	}
	if provider != "" {
		q = q.Where("s.billing_provider = ?", provider)
	}
	var rows []OversightRow
	err := q.Scan(&rows).Error
	return rows, err
}
