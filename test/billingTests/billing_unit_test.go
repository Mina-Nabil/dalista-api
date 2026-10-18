package billingTests

import (
	"testing"
	"time"

	"dalista/dalista-be/api/resources"
	billingModels "dalista/dalista-be/app/repository/models/billingModels"
	billingServices "dalista/dalista-be/app/services/billingServices"
	"dalista/dalista-be/integrations/appstore"
	"dalista/dalista-be/integrations/playbilling"
)

const gib = int64(1) << 30

func TestTierLimits(t *testing.T) {
	cases := []struct {
		tier    billingModels.Tier
		storage int64
		cap     int
	}{
		{billingModels.TierFree, 2 * gib, 5},
		{billingModels.TierBasic, 5 * gib, 15},
		{billingModels.TierUnlimited, 50 * gib, billingServices.Unlimited},
	}
	for _, tc := range cases {
		l := billingServices.LimitsFor(tc.tier)
		if l.StorageBytes != tc.storage || l.PerListUserCap != tc.cap {
			t.Errorf("%s limits = %+v, want storage %d cap %d", tc.tier, l, tc.storage, tc.cap)
		}
	}
	// Unknown tier falls back to free.
	if billingServices.LimitsFor("bogus").StorageBytes != 2*gib {
		t.Error("unknown tier should fall back to free limits")
	}
}

func TestTierForProduct(t *testing.T) {
	if billingServices.TierForProduct("dalista.unlimited.yearly") != billingModels.TierUnlimited {
		t.Error("unlimited product should map to unlimited tier")
	}
	if billingServices.TierForProduct("dalista.basic.monthly") != billingModels.TierBasic {
		t.Error("basic product should map to basic tier")
	}
	// Unknown paid product must never silently grant unlimited — defaults to basic.
	if billingServices.TierForProduct("mystery.product") != billingModels.TierBasic {
		t.Error("unknown product should default to basic")
	}
}

func TestSubscriptionResourceDefaultsToFree(t *testing.T) {
	r := resources.SubscriptionFromModel(nil)
	if r.Tier != "free" || !r.Entitled {
		t.Errorf("nil subscription should render entitled free tier, got %+v", r)
	}
	exp := time.Now().Add(24 * time.Hour)
	sub := &billingModels.Subscription{
		Tier: billingModels.TierBasic, Status: billingModels.StatusActive,
		BillingProvider: billingModels.ProviderAppleIAP, CurrentPeriodEnd: &exp,
	}
	r2 := resources.SubscriptionFromModel(sub)
	if r2.Tier != "basic" || !r2.Entitled || r2.CurrentPeriodEnd == nil {
		t.Errorf("active basic subscription mapped wrong: %+v", r2)
	}
}

func TestAppleTransactionHelpers(t *testing.T) {
	zero := appstore.TransactionInfo{}
	if zero.ExpiresAt() != nil || zero.Revoked() {
		t.Error("zero-value transaction should have nil expiry and not be revoked")
	}
	txn := appstore.TransactionInfo{ExpiresDate: time.Now().UnixMilli(), RevocationDate: 123}
	if txn.ExpiresAt() == nil {
		t.Error("expected non-nil expiry")
	}
	if !txn.Revoked() {
		t.Error("expected revoked when RevocationDate is set")
	}
}

func TestRTDNStatusMapping(t *testing.T) {
	if ent, ok := playbilling.StatusForNotificationType(playbilling.RTDNRenewed); !ok || !ent {
		t.Error("renewed should be recognized + entitled")
	}
	if ent, ok := playbilling.StatusForNotificationType(playbilling.RTDNExpired); !ok || ent {
		t.Error("expired should be recognized + not entitled")
	}
	if _, ok := playbilling.StatusForNotificationType(9999); ok {
		t.Error("unknown type should be unrecognized")
	}
}
