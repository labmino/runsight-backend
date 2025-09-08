package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PairingSession struct {
	ID        string     `json:"id" gorm:"type:varchar(50);primary_key"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Code      string     `json:"code" gorm:"type:varchar(6);not null;index" validate:"required,len=6"`
	DeviceID  string     `json:"device_id,omitempty" gorm:"type:varchar(50);index"`
	Status    string     `json:"status" gorm:"type:varchar(20);default:'pending';index" validate:"oneof=pending paired expired"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	PairedAt  *time.Time `json:"paired_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	User   User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type PairingRequest struct {}

type PairingResponse struct {
	Code              string `json:"code"`
	SessionID         string `json:"session_id"`
	ExpiresAt         string `json:"expires_at"`
	ExpiresInSeconds  int    `json:"expires_in_seconds"`
}

type PairingStatusResponse struct {
	Paired           bool                    `json:"paired"`
	Expired          bool                    `json:"expired"`
	RemainingSeconds int                     `json:"remaining_seconds,omitempty"`
	Device           *PairingDeviceInfo      `json:"device,omitempty"`
}

type PairingDeviceInfo struct {
	DeviceID         string    `json:"device_id"`
	DeviceName       string    `json:"device_name"`
	DeviceType       string    `json:"device_type"`
	FirmwareVersion  string    `json:"firmware_version"`
	PairedAt         time.Time `json:"paired_at"`
}

const (
	PairingStatusPending = "pending"
	PairingStatusPaired  = "paired" 
	PairingStatusExpired = "expired"
)

const PairingCodeTTL = 5 * time.Minute

func (p *PairingSession) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = "pair_" + uuid.New().String()[:8]
	}
	if p.Status == "" {
		p.Status = PairingStatusPending
	}
	if p.ExpiresAt.IsZero() {
		p.ExpiresAt = time.Now().Add(PairingCodeTTL)
	}
	p.CreatedAt = time.Now()
	return nil
}

func (p *PairingSession) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

func (p *PairingSession) RemainingSeconds() int {
	remaining := time.Until(p.ExpiresAt).Seconds()
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}