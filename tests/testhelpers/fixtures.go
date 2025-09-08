package testhelpers

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateTestUser(db *gorm.DB, email string) *TestUser {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)
	
	user := &TestUser{
		ID:           uuid.New().String(),
		FullName:     "Test User",
		Email:        email,
		Phone:        "+1234567890",
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	db.Create(user)
	return user
}

func CreateTestDevice(db *gorm.DB, userID string, deviceID string) *TestDevice {
	device := &TestDevice{
		ID:               uuid.New().String(),
		DeviceID:         deviceID,
		UserID:           userID,
		DeviceName:       "Test Device",
		DeviceType:       "smart_glasses",
		FirmwareVersion:  "1.0.0",
		HardwareVersion:  "1.0.0",
		MACAddress:       "AA:BB:CC:DD:EE:FF",
		DeviceToken:      "test_device_token_" + deviceID,
		IsActive:         true,
		BatteryLevel:     nil,
		LastSyncAt:       nil,
		PairedAt:         time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	
	db.Create(device)
	return device
}

func CreateTestRun(db *gorm.DB, userID string, deviceID string) *TestRun {
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := startTime.Add(30 * time.Minute)
	duration := 1800
	distance := 5000.0
	avgSpeed := 10.0
	maxSpeed := 15.0
	calories := 300
	steps := 6500

	run := &TestRun{
		ID:              uuid.New().String(),
		UserID:          userID,
		DeviceID:        deviceID,
		SessionID:       "test_session_" + uuid.New().String(),
		Title:           "Test Run",
		Notes:           "Test run notes",
		StartedAt:       startTime,
		EndedAt:         &endTime,
		DurationSeconds: &duration,
		DistanceMeters:  &distance,
		AvgSpeedKmh:     &avgSpeed,
		MaxSpeedKmh:     &maxSpeed,
		CaloriesBurned:  &calories,
		StepsCount:      &steps,
		StartLatitude:   nil,
		StartLongitude:  nil,
		EndLatitude:     nil,
		EndLongitude:    nil,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	db.Create(run)
	return run
}

func CreateTestAIMetrics(db *gorm.DB, runID string) *TestAIMetrics {
	totalFrames := 5400
	obstacles := 12
	warnings := 3
	deviations := 2
	accuracy := 0.95
	avgInference := 25.5
	maxInference := 45.2
	minInference := 15.8

	aiMetrics := &TestAIMetrics{
		ID:                       uuid.New().String(),
		RunID:                    runID,
		TotalFramesProcessed:     &totalFrames,
		TotalObstaclesDetected:   &obstacles,
		TotalWarningsIssued:      &warnings,
		LaneDeviationsCount:      &deviations,
		LaneKeepingAccuracy:      &accuracy,
		AvgInferenceTimeMs:       &avgInference,
		MaxInferenceTimeMs:       &maxInference,
		MinInferenceTimeMs:       &minInference,
		CreatedAt:                time.Now(),
	}
	
	db.Create(aiMetrics)
	return aiMetrics
}