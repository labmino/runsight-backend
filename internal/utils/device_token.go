package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"gorm.io/gorm"
	
	"github.com/labmino/runsight-backend/internal/models"
)

func GenerateDeviceToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ValidateDeviceToken(db *gorm.DB, tokenString string) (*models.Device, error) {
	if tokenString == "" {
		return nil, errors.New("device token is required")
	}

	var device models.Device
	err := db.Where("device_token = ? AND is_active = ?", tokenString, true).First(&device).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or inactive device token")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &device, nil
}

func GeneratePairingCode() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	code := fmt.Sprintf("%06d", int(bytes[0])<<16|int(bytes[1])<<8|int(bytes[2])%1000000)
	if len(code) > 6 {
		code = code[:6]
	}
	for len(code) < 6 {
		code = "0" + code
	}
	return code
}