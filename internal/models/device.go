package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DeviceID         string     `json:"device_id" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required"`
	UserID           uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	DeviceName       string     `json:"device_name" gorm:"type:varchar(100)" validate:"omitempty,max=100"`
	DeviceType       string     `json:"device_type" gorm:"type:varchar(50);not null" validate:"required"`
	FirmwareVersion  string     `json:"firmware_version" gorm:"type:varchar(20)" validate:"omitempty,max=20"`
	HardwareVersion  string     `json:"hardware_version" gorm:"type:varchar(20)" validate:"omitempty,max=20"`
	MACAddress       string     `json:"mac_address" gorm:"type:varchar(17)" validate:"omitempty,mac"`
	DeviceToken      string     `json:"-" gorm:"type:text"`
	IsActive         bool       `json:"is_active" gorm:"default:true"`
	BatteryLevel     *int       `json:"battery_level,omitempty" validate:"omitempty,min=0,max=100"`
	LastSyncAt       *time.Time `json:"last_sync_at,omitempty"`
	PairedAt         time.Time  `json:"paired_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	User User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Runs []Run `json:"runs,omitempty" gorm:"foreignKey:DeviceID;references:DeviceID"`
}

type DeviceRegisterRequest struct {
	Code            string `json:"code" validate:"required,len=6"`
	DeviceID        string `json:"device_id" validate:"required,max=50"`
	DeviceType      string `json:"device_type" validate:"required,max=50"`
	FirmwareVersion string `json:"firmware_version,omitempty" validate:"omitempty,max=20"`
	HardwareVersion string `json:"hardware_version,omitempty" validate:"omitempty,max=20"`
	MACAddress      string `json:"mac_address,omitempty" validate:"omitempty,mac"`
}

type DeviceStatusRequest struct {
	DeviceID            string     `json:"device_id" validate:"required"`
	BatteryLevel        *int       `json:"battery_level,omitempty" validate:"omitempty,min=0,max=100"`
	StorageAvailableMB  *int       `json:"storage_available_mb,omitempty" validate:"omitempty,min=0"`
	FirmwareVersion     string     `json:"firmware_version,omitempty" validate:"omitempty,max=20"`
	ErrorCount          *int       `json:"error_count,omitempty" validate:"omitempty,min=0"`
	LastRunAt           *time.Time `json:"last_run_at,omitempty"`
}

type DeviceConfig struct {
	UploadIntervalSeconds int  `json:"upload_interval_seconds"`
	BatchSize            int  `json:"batch_size"`
	CompressionEnabled   bool `json:"compression_enabled"`
}

func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	d.PairedAt = time.Now()
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	return nil
}

func (d *Device) BeforeUpdate(tx *gorm.DB) error {
	d.UpdatedAt = time.Now()
	return nil
}