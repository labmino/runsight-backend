package models

import (
	"time"

	"gorm.io/gorm"
)

type Run struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	UserID          int64      `json:"user_id" gorm:"not null"`
	Title           string     `json:"title,omitempty" gorm:"type:varchar(100)" validate:"omitempty,max=100"`
	Notes           string     `json:"notes,omitempty" gorm:"type:text"`
	StartedAt       time.Time  `json:"started_at" gorm:"not null" validate:"required"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" validate:"omitempty,min=1"`
	DistanceMeters  *float64   `json:"distance_meters,omitempty" gorm:"type:decimal(10,2)" validate:"omitempty,min=0"`
	AvgSpeedKmh     *float64   `json:"avg_speed_kmh,omitempty" gorm:"type:decimal(8,2)" validate:"omitempty,min=0"`
	MaxSpeedKmh     *float64   `json:"max_speed_kmh,omitempty" gorm:"type:decimal(8,2)" validate:"omitempty,min=0"`
	StartLatitude   *float64   `json:"start_latitude,omitempty" gorm:"type:decimal(10,8)" validate:"omitempty,latitude"`
	StartLongitude  *float64   `json:"start_longitude,omitempty" gorm:"type:decimal(11,8)" validate:"omitempty,longitude"`
	EndLatitude     *float64   `json:"end_latitude,omitempty" gorm:"type:decimal(10,8)" validate:"omitempty,latitude"`
	EndLongitude    *float64   `json:"end_longitude,omitempty" gorm:"type:decimal(11,8)" validate:"omitempty,longitude"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type RunCreateRequest struct {
	Title           string     `json:"title,omitempty" validate:"omitempty,max=100"`
	Notes           string     `json:"notes,omitempty"`
	StartedAt       time.Time  `json:"started_at" validate:"required"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" validate:"omitempty,min=1"`
	DistanceMeters  *float64   `json:"distance_meters,omitempty" validate:"omitempty,min=0"`
	AvgSpeedKmh     *float64   `json:"avg_speed_kmh,omitempty" validate:"omitempty,min=0"`
	MaxSpeedKmh     *float64   `json:"max_speed_kmh,omitempty" validate:"omitempty,min=0"`
	StartLatitude   *float64   `json:"start_latitude,omitempty" validate:"omitempty,latitude"`
	StartLongitude  *float64   `json:"start_longitude,omitempty" validate:"omitempty,longitude"`
	EndLatitude     *float64   `json:"end_latitude,omitempty" validate:"omitempty,latitude"`
	EndLongitude    *float64   `json:"end_longitude,omitempty" validate:"omitempty,longitude"`
}

type RunUpdateRequest struct {
	Title           *string    `json:"title,omitempty" validate:"omitempty,max=100"`
	Notes           *string    `json:"notes,omitempty"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" validate:"omitempty,min=1"`
	DistanceMeters  *float64   `json:"distance_meters,omitempty" validate:"omitempty,min=0"`
	AvgSpeedKmh     *float64   `json:"avg_speed_kmh,omitempty" validate:"omitempty,min=0"`
	MaxSpeedKmh     *float64   `json:"max_speed_kmh,omitempty" validate:"omitempty,min=0"`
	EndLatitude     *float64   `json:"end_latitude,omitempty" validate:"omitempty,latitude"`
	EndLongitude    *float64   `json:"end_longitude,omitempty" validate:"omitempty,longitude"`
}

func (r *Run) BeforeCreate(tx *gorm.DB) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Run) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}