package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIMetrics struct {
	ID                       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RunID                    uuid.UUID `json:"run_id" gorm:"type:uuid;not null;index"`
	TotalFramesProcessed     *int      `json:"total_frames_processed,omitempty" validate:"omitempty,min=0"`
	TotalObstaclesDetected   *int      `json:"total_obstacles_detected,omitempty" validate:"omitempty,min=0"`
	TotalWarningsIssued      *int      `json:"total_warnings_issued,omitempty" validate:"omitempty,min=0"`
	LaneDeviationsCount      *int      `json:"lane_deviations_count,omitempty" validate:"omitempty,min=0"`
	LaneKeepingAccuracy      *float64  `json:"lane_keeping_accuracy,omitempty" gorm:"type:decimal(3,2)" validate:"omitempty,min=0,max=1"`
	AvgInferenceTimeMs       *float64  `json:"avg_inference_time_ms,omitempty" gorm:"type:decimal(6,2)" validate:"omitempty,min=0"`
	MaxInferenceTimeMs       *float64  `json:"max_inference_time_ms,omitempty" gorm:"type:decimal(6,2)" validate:"omitempty,min=0"`
	MinInferenceTimeMs       *float64  `json:"min_inference_time_ms,omitempty" gorm:"type:decimal(6,2)" validate:"omitempty,min=0"`
	CreatedAt                time.Time `json:"created_at"`

	Run Run `json:"run,omitempty" gorm:"foreignKey:RunID"`
}

type AIMetricsRequest struct {
	TotalFrames          *int     `json:"total_frames,omitempty" validate:"omitempty,min=0"`
	ObstaclesDetected    *int     `json:"obstacles_detected,omitempty" validate:"omitempty,min=0"`
	WarningsIssued       *int     `json:"warnings_issued,omitempty" validate:"omitempty,min=0"`
	LaneDeviations       *int     `json:"lane_deviations,omitempty" validate:"omitempty,min=0"`
	LaneKeepingAccuracy  *float64 `json:"lane_keeping_accuracy,omitempty" validate:"omitempty,min=0,max=1"`
	AvgInferenceMs       *float64 `json:"avg_inference_ms,omitempty" validate:"omitempty,min=0"`
	MaxInferenceMs       *float64 `json:"max_inference_ms,omitempty" validate:"omitempty,min=0"`
	MinInferenceMs       *float64 `json:"min_inference_ms,omitempty" validate:"omitempty,min=0"`
}

func (req *AIMetricsRequest) ToModel(runID uuid.UUID) *AIMetrics {
	return &AIMetrics{
		RunID:                    runID,
		TotalFramesProcessed:     req.TotalFrames,
		TotalObstaclesDetected:   req.ObstaclesDetected,
		TotalWarningsIssued:      req.WarningsIssued,
		LaneDeviationsCount:      req.LaneDeviations,
		LaneKeepingAccuracy:      req.LaneKeepingAccuracy,
		AvgInferenceTimeMs:       req.AvgInferenceMs,
		MaxInferenceTimeMs:       req.MaxInferenceMs,
		MinInferenceTimeMs:       req.MinInferenceMs,
	}
}

func (ai *AIMetrics) BeforeCreate(tx *gorm.DB) error {
	if ai.ID == uuid.Nil {
		ai.ID = uuid.New()
	}
	ai.CreatedAt = time.Now()
	return nil
}