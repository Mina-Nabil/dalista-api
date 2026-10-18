// Command syncpermissions upserts the permission registry (permissions +
// role_permissions tables) from the Go-defined registry. Run after migrations,
// idempotently, on every deploy (see the CD workflow's migrate/sync steps).
package main

import (
	"fmt"
	"time"

	"dalista/dalista-be/api/server/logger"
	"dalista/dalista-be/app/repository"
	"dalista/dalista-be/app/services"
	authServices "dalista/dalista-be/app/services/authServices"
	"dalista/dalista-be/config"
)

func main() {
	cfg := config.GetConfig()
	if err := logger.Init(logger.InitOptions{DSN: cfg.SentryDSN, Environment: cfg.AppEnv}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	defer logger.Flush(2 * time.Second)

	fmt.Println("sync-permissions: starting…")

	_ = repository.GetDB()

	allPerms := services.AllPermissions()
	roleDefaults := services.RoleDefaults()

	if err := authServices.SyncPermissions(allPerms, roleDefaults); err != nil {
		logger.Fatal(err, "sync-permissions failed")
	}

	fmt.Printf("sync-permissions: done — %d permissions, %d roles\n", len(allPerms), len(roleDefaults))
}
