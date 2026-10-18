package repository

import (
	"fmt"

	"dalista/dalista-be/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// GetDB returns the process-wide GORM handle, lazily opening the connection on
// first use. Services call this for all DB access (never open their own).
func GetDB() *gorm.DB {
	if db == nil {
		if err := initializeDB(); err != nil {
			panic(err)
		}
	}
	return db
}

// SetDB overrides the handle — used by tests to inject a migrated test database.
func SetDB(dbInstance *gorm.DB) {
	db = dbInstance
}

// Ping verifies database connectivity without panicking, so callers like the
// health probe can report "down" gracefully. It lazily opens the connection on
// first use, matching GetDB.
func Ping() error {
	if db == nil {
		if err := initializeDB(); err != nil {
			return err
		}
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func initializeDB() error {
	cfg := config.GetConfig()
	dbURL := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	dbInstance, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return err
	}
	SetDB(dbInstance)
	return nil
}
