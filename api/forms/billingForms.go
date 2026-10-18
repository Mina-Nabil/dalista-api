package forms

// AppleVerifyForm carries the transaction identifier the client received from
// StoreKit; the backend verifies it against the App Store Server API.
type AppleVerifyForm struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

// GoogleVerifyForm carries the Play Billing purchase token + subscription
// (product) id; the backend verifies it against the Play Developer API.
type GoogleVerifyForm struct {
	SubscriptionID string `json:"subscription_id" binding:"required" example:"dalista.basic.monthly"`
	PurchaseToken  string `json:"purchase_token" binding:"required"`
}
