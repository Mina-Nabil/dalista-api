package resources

import (
	"time"

	billingModels "dalista/dalista-be/app/repository/models/billingModels"
)

// SubscriptionResource is the user-facing view of a subscription. A nil model
// renders the implicit free tier.
type SubscriptionResource struct {
	Tier             string  `json:"tier"`
	Status           string  `json:"status"`
	BillingProvider  *string `json:"billing_provider"`
	ProductID        *string `json:"product_id"`
	CurrentPeriodEnd *string `json:"current_period_end"`
	Entitled         bool    `json:"entitled"`
}

// SubscriptionFromModel maps a subscription (or nil = free) to its resource.
func SubscriptionFromModel(s *billingModels.Subscription) SubscriptionResource {
	if s == nil {
		return SubscriptionResource{
			Tier:     string(billingModels.TierFree),
			Status:   string(billingModels.StatusActive),
			Entitled: true,
		}
	}
	provider := string(s.BillingProvider)
	res := SubscriptionResource{
		Tier:            string(s.Tier),
		Status:          string(s.Status),
		BillingProvider: &provider,
		ProductID:       s.ProductID,
		Entitled:        s.Entitled(),
	}
	if s.CurrentPeriodEnd != nil {
		t := s.CurrentPeriodEnd.UTC().Format(time.RFC3339)
		res.CurrentPeriodEnd = &t
	}
	return res
}
