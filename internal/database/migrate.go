package database

import (
	"gorm.io/gorm"
	"github.com/labmino/runsight-backend/internal/models"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.PairingSession{},
		&models.Run{},
		&models.AIMetrics{},
	)
}