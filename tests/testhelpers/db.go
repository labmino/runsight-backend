package testhelpers

import (
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel: logger.Silent,
			},
		),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect to test database")
	}

	db.Exec("PRAGMA foreign_keys = ON")

	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		panic("failed to migrate TestUser model: " + err.Error())
	}

	err = db.AutoMigrate(&TestDevice{})
	if err != nil {
		panic("failed to migrate TestDevice model: " + err.Error())
	}

	err = db.AutoMigrate(&TestRun{})
	if err != nil {
		panic("failed to migrate TestRun model: " + err.Error())
	}

	err = db.AutoMigrate(&TestAIMetrics{})
	if err != nil {
		panic("failed to migrate TestAIMetrics model: " + err.Error())
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func TruncateTables(db *gorm.DB) {
	db.Exec("DELETE FROM test_ai_metrics")
	db.Exec("DELETE FROM test_runs")
	db.Exec("DELETE FROM test_devices")
	db.Exec("DELETE FROM test_users")
}