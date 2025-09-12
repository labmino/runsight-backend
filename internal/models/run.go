package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Run struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	DeviceID        string     `json:"device_id" gorm:"type:varchar(50);index" validate:"required"`
	SessionID       string     `json:"session_id" gorm:"type:varchar(100);uniqueIndex;not null" validate:"required"`
	Title           string     `json:"title,omitempty" gorm:"type:varchar(100)" validate:"omitempty,max=100"`
	Notes           string     `json:"notes,omitempty" gorm:"type:text"`
	StartedAt       time.Time  `json:"started_at" gorm:"not null" validate:"required"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" validate:"omitempty,min=1"`
	DistanceMeters  *float64   `json:"distance_meters,omitempty" gorm:"type:decimal(10,2)" validate:"omitempty,min=0"`
	AvgSpeedKmh     *float64   `json:"avg_speed_kmh,omitempty" gorm:"type:decimal(8,2)" validate:"omitempty,min=0"`
	MaxSpeedKmh     *float64   `json:"max_speed_kmh,omitempty" gorm:"type:decimal(8,2)" validate:"omitempty,min=0"`
	CaloriesBurned  *int       `json:"calories_burned,omitempty" validate:"omitempty,min=0"`
	StepsCount      *int       `json:"steps_count,omitempty" validate:"omitempty,min=0"`
	StartLatitude   *float64   `json:"start_latitude,omitempty" gorm:"type:decimal(10,8)" validate:"omitempty,latitude"`
	StartLongitude  *float64   `json:"start_longitude,omitempty" gorm:"type:decimal(11,8)" validate:"omitempty,longitude"`
	EndLatitude     *float64   `json:"end_latitude,omitempty" gorm:"type:decimal(10,8)" validate:"omitempty,latitude"`
	EndLongitude    *float64   `json:"end_longitude,omitempty" gorm:"type:decimal(11,8)" validate:"omitempty,longitude"`
	RouteData       *string    `json:"route_data,omitempty" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type RunCreateRequest struct {
	DeviceID        string     `json:"device_id" validate:"required"`
	SessionID       string     `json:"session_id" validate:"required"`
	Title           string     `json:"title,omitempty" validate:"omitempty,max=100"`
	Notes           string     `json:"notes,omitempty"`
	StartedAt       time.Time  `json:"started_at" validate:"required"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" validate:"omitempty,min=1"`
	DistanceMeters  *float64   `json:"distance_meters,omitempty" validate:"omitempty,min=0"`
	AvgSpeedKmh     *float64   `json:"avg_speed_kmh,omitempty" validate:"omitempty,min=0"`
	MaxSpeedKmh     *float64   `json:"max_speed_kmh,omitempty" validate:"omitempty,min=0"`
	CaloriesBurned  *int       `json:"calories_burned,omitempty" validate:"omitempty,min=0"`
	StepsCount      *int       `json:"steps_count,omitempty" validate:"omitempty,min=0"`
	StartLatitude   *float64   `json:"start_latitude,omitempty" validate:"omitempty,latitude"`
	StartLongitude  *float64   `json:"start_longitude,omitempty" validate:"omitempty,longitude"`
	EndLatitude     *float64   `json:"end_latitude,omitempty" validate:"omitempty,latitude"`
	EndLongitude    *float64   `json:"end_longitude,omitempty" validate:"omitempty,longitude"`
	Waypoints       []WaypointData `json:"waypoints,omitempty" validate:"omitempty,dive"`
}

type RunUpdateRequest struct {
	Title           *string    `json:"title,omitempty" validate:"omitempty,max=100"`
	Notes           *string    `json:"notes,omitempty"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" validate:"omitempty,min=1"`
	DistanceMeters  *float64   `json:"distance_meters,omitempty" validate:"omitempty,min=0"`
	AvgSpeedKmh     *float64   `json:"avg_speed_kmh,omitempty" validate:"omitempty,min=0"`
	MaxSpeedKmh     *float64   `json:"max_speed_kmh,omitempty" validate:"omitempty,min=0"`
	CaloriesBurned  *int       `json:"calories_burned,omitempty" validate:"omitempty,min=0"`
	StepsCount      *int       `json:"steps_count,omitempty" validate:"omitempty,min=0"`
	EndLatitude     *float64   `json:"end_latitude,omitempty" validate:"omitempty,latitude"`
	EndLongitude    *float64   `json:"end_longitude,omitempty" validate:"omitempty,longitude"`
}

func (r *Run) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Run) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

type WaypointData struct {
	Latitude  float64   `json:"lat" validate:"required,latitude"`
	Longitude float64   `json:"lng" validate:"required,longitude"`
	Speed     *float64  `json:"speed,omitempty" validate:"omitempty,min=0"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}