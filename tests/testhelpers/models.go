package testhelpers

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Test-specific models that are SQLite compatible
type TestUser struct {
	ID           string    `json:"id" gorm:"primary_key"`
	FullName     string    `json:"full_name" gorm:"not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	Phone        string    `json:"phone,omitempty"`
	PasswordHash string    `json:"-" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TestDevice struct {
	ID               string     `json:"id" gorm:"primary_key"`
	DeviceID         string     `json:"device_id" gorm:"uniqueIndex;not null"`
	UserID           string     `json:"user_id" gorm:"not null"`
	DeviceName       string     `json:"device_name"`
	DeviceType       string     `json:"device_type" gorm:"not null"`
	FirmwareVersion  string     `json:"firmware_version"`
	HardwareVersion  string     `json:"hardware_version"`
	MACAddress       string     `json:"mac_address"`
	DeviceToken      string     `json:"-"`
	IsActive         bool       `json:"is_active" gorm:"default:true"`
	BatteryLevel     *int       `json:"battery_level,omitempty"`
	LastSyncAt       *time.Time `json:"last_sync_at,omitempty"`
	PairedAt         time.Time  `json:"paired_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TestRun struct {
	ID              string     `json:"id" gorm:"primary_key"`
	UserID          string     `json:"user_id" gorm:"not null"`
	DeviceID        string     `json:"device_id"`
	SessionID       string     `json:"session_id" gorm:"uniqueIndex;not null"`
	Title           string     `json:"title,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	StartedAt       time.Time  `json:"started_at" gorm:"not null"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	DistanceMeters  *float64   `json:"distance_meters,omitempty"`
	AvgSpeedKmh     *float64   `json:"avg_speed_kmh,omitempty"`
	MaxSpeedKmh     *float64   `json:"max_speed_kmh,omitempty"`
	CaloriesBurned  *int       `json:"calories_burned,omitempty"`
	StepsCount      *int       `json:"steps_count,omitempty"`
	StartLatitude   *float64   `json:"start_latitude,omitempty"`
	StartLongitude  *float64   `json:"start_longitude,omitempty"`
	EndLatitude     *float64   `json:"end_latitude,omitempty"`
	EndLongitude    *float64   `json:"end_longitude,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TestAIMetrics struct {
	ID                       string   `json:"id" gorm:"primary_key"`
	RunID                    string   `json:"run_id" gorm:"not null"`
	TotalFramesProcessed     *int     `json:"total_frames_processed,omitempty"`
	TotalObstaclesDetected   *int     `json:"total_obstacles_detected,omitempty"`
	TotalWarningsIssued      *int     `json:"total_warnings_issued,omitempty"`
	LaneDeviationsCount      *int     `json:"lane_deviations_count,omitempty"`
	LaneKeepingAccuracy      *float64 `json:"lane_keeping_accuracy,omitempty"`
	AvgInferenceTimeMs       *float64 `json:"avg_inference_time_ms,omitempty"`
	MaxInferenceTimeMs       *float64 `json:"max_inference_time_ms,omitempty"`
	MinInferenceTimeMs       *float64 `json:"min_inference_time_ms,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
}

func (t *TestUser) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}

func (t *TestDevice) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.PairedAt = time.Now()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}

func (t *TestRun) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}

func (t *TestAIMetrics) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now()
	return nil
}