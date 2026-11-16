// Command devseed creates a known local-dev account with sample data so the
// mobile app (and manual testing) has something to render immediately. It is
// idempotent: re-running reuses the existing account and skips sample data if it
// is already present. DEV ONLY — it refuses to run outside a staging/dev env.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/app/repository"
	authServices "dalista/dalista-be/app/services/authServices"
	listServices "dalista/dalista-be/app/services/listServices"
	"dalista/dalista-be/config"
)

const (
	seedEmail    = "dev@dalista.app"
	seedPassword = "password123"
)

func main() {
	cfg := config.GetConfig()
	if err := logger.Init(logger.InitOptions{DSN: cfg.SentryDSN, Environment: cfg.AppEnv}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	defer logger.Flush(2 * time.Second)

	if !cfg.IsStagingEnvironment() {
		fmt.Println("devseed: refusing to run outside a development/staging environment")
		return
	}

	ctx := context.Background()
	_ = repository.GetDB()

	// 1. Account (idempotent).
	user, _, err := authServices.RegisterEmail(ctx, seedEmail, seedPassword)
	if errors.Is(err, authServices.ErrEmailTaken) {
		user, err = authServices.GetUserByEmail(ctx, seedEmail)
	}
	if err != nil || user == nil {
		logger.Fatal(err, "devseed: could not create/fetch seed user")
	}
	fmt.Printf("devseed: account ready — %s / %s  (id %s)\n", seedEmail, seedPassword, user.ID)

	// 2. Sample data — skip if the account already has lists.
	existing, err := listServices.ListLists(ctx, user.ID, nil)
	if err != nil {
		logger.Fatal(err, "devseed: could not check existing lists")
	}
	if len(existing) > 0 {
		fmt.Printf("devseed: %d list(s) already present — skipping sample data\n", len(existing))
		return
	}

	cat, err := listServices.CreateCategory(ctx, user.ID, "Groceries")
	if err != nil {
		logger.Fatal(err, "devseed: category")
	}
	group, err := listServices.CreateGroupedList(ctx, user.ID, "Home")
	if err != nil {
		logger.Fatal(err, "devseed: grouped list")
	}

	truePtr := true
	groupID := group.ID
	catID := cat.ID

	// Periodic weekly-groceries list, nested under the Home group.
	weekly, err := listServices.CreateList(ctx, user.ID, listServices.ListInput{
		Name: "Weekly groceries", GroupedListID: &groupID, CategoryID: &catID, IsPeriodic: &truePtr,
	})
	if err != nil {
		logger.Fatal(err, "devseed: weekly list")
	}
	for _, name := range []string{"Milk", "Eggs", "Bread", "Coffee"} {
		price := 9.5
		if _, err := listServices.CreateItem(ctx, user.ID, weekly.ID, listServices.ItemInput{Name: name, Price: &price}); err != nil {
			logger.Fatal(err, "devseed: weekly item")
		}
	}

	// A public electronics wishlist with a couple of items.
	wishlist, err := listServices.CreateList(ctx, user.ID, listServices.ListInput{
		Name: "Electronics wishlist", IsPublic: &truePtr,
	})
	if err != nil {
		logger.Fatal(err, "devseed: wishlist")
	}
	for _, it := range []struct {
		name  string
		brand string
		price float64
	}{{"Headphones", "Sony", 299.0}, {"Mechanical keyboard", "Keychron", 110.0}} {
		b := it.brand
		p := it.price
		if _, err := listServices.CreateItem(ctx, user.ID, wishlist.ID, listServices.ItemInput{Name: it.name, Brand: &b, Price: &p}); err != nil {
			logger.Fatal(err, "devseed: wishlist item")
		}
	}

	fmt.Println("devseed: sample data created —")
	fmt.Printf("  • category 'Groceries', group 'Home'\n")
	fmt.Printf("  • periodic list 'Weekly groceries' (4 items)\n")
	fmt.Printf("  • public list 'Electronics wishlist' (2 items)\n")
	fmt.Println("devseed: done.")
}
