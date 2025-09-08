package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/labmino/runsight-backend/internal/handlers"
	"github.com/labmino/runsight-backend/tests/testhelpers"
)

type MobileHandlerTestSuite struct {
	suite.Suite
	handler *handlers.MobileHandler
	router  *gin.Engine
	cleanup func()
}

func (suite *MobileHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	db, cleanup := testhelpers.SetupTestDB()
	suite.cleanup = cleanup
	suite.handler = handlers.NewMobileHandler(db)
	suite.router = gin.New()
	
	// Setup routes
	mobile := suite.router.Group("/mobile")
	mobile.Use(suite.mockAuthMiddleware())
	{
		mobile.GET("/runs", suite.handler.ListRuns)
		mobile.GET("/runs/:run_id", suite.handler.GetRun)
		mobile.PATCH("/runs/:run_id", suite.handler.UpdateRunNotes)
		mobile.GET("/stats", suite.handler.GetStats)
		mobile.GET("/devices", suite.handler.GetDevices)
		mobile.DELETE("/devices/:device_id", suite.handler.RemoveDevice)
		mobile.POST("/pairing/request", suite.handler.RequestPairingCode)
		mobile.GET("/pairing/:session_id/status", suite.handler.CheckPairingStatus)
	}
}

func (suite *MobileHandlerTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *MobileHandlerTestSuite) SetupTest() {
	testhelpers.TruncateTables(suite.handler.DB())
}

func (suite *MobileHandlerTestSuite) mockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock user authentication
		testUserID := uuid.New()
		c.Set("user_id", testUserID)
		c.Set("user_email", "test@example.com")
		c.Next()
	}
}

func (suite *MobileHandlerTestSuite) TestListRuns() {
	// Create test data
	user := testhelpers.CreateTestUser(suite.handler.DB(), "test@example.com")
	device := testhelpers.CreateTestDevice(suite.handler.DB(), user.ID, "TEST_DEVICE")
	testhelpers.CreateTestRun(suite.handler.DB(), user.ID, device.DeviceID)
	
	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	// Test successful list
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mobile/runs?page=1&limit=10", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
}

func (suite *MobileHandlerTestSuite) TestGetRun() {
	// Create test data
	user := testhelpers.CreateTestUser(suite.handler.DB(), "test@example.com")
	device := testhelpers.CreateTestDevice(suite.handler.DB(), user.ID, "TEST_DEVICE")
	run := testhelpers.CreateTestRun(suite.handler.DB(), user.ID, device.DeviceID)
	testhelpers.CreateTestAIMetrics(suite.handler.DB(), run.ID)

	// Mock auth middleware for this test
	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	// Test successful get
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/mobile/runs/%s", run.ID), nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), run.ID.String(), data["id"])
	assert.NotNil(suite.T(), data["ai_metrics"])
}

func (suite *MobileHandlerTestSuite) TestGetRunNotFound() {
	testUserID := uuid.New()
	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})

	nonExistentID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/mobile/runs/%s", nonExistentID), nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *MobileHandlerTestSuite) TestUpdateRunNotes() {
	// Create test data
	user := testhelpers.CreateTestUser(suite.handler.DB(), "test@example.com")
	device := testhelpers.CreateTestDevice(suite.handler.DB(), user.ID, "TEST_DEVICE")
	run := testhelpers.CreateTestRun(suite.handler.DB(), user.ID, device.DeviceID)

	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	// Test successful update
	updateData := map[string]interface{}{
		"title": "Updated Title",
		"notes": "Updated notes",
	}
	jsonData, _ := json.Marshal(updateData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/mobile/runs/%s", run.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Updated Title", data["title"])
	assert.Equal(suite.T(), "Updated notes", data["notes"])
}

func (suite *MobileHandlerTestSuite) TestGetStats() {
	// Create test data
	user := testhelpers.CreateTestUser(suite.handler.DB(), "test@example.com")
	device := testhelpers.CreateTestDevice(suite.handler.DB(), user.ID, "TEST_DEVICE")
	testhelpers.CreateTestRun(suite.handler.DB(), user.ID, device.DeviceID)
	testhelpers.CreateTestRun(suite.handler.DB(), user.ID, device.DeviceID)

	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	// Test successful stats retrieval
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mobile/stats", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(2), data["total_runs"])
	assert.Equal(suite.T(), float64(10000), data["total_distance_meters"])
}

func (suite *MobileHandlerTestSuite) TestGetDevices() {
	// Create test data
	user := testhelpers.CreateTestUser(suite.handler.DB(), "test@example.com")
	testhelpers.CreateTestDevice(suite.handler.DB(), user.ID, "TEST_DEVICE_1")
	testhelpers.CreateTestDevice(suite.handler.DB(), user.ID, "TEST_DEVICE_2")

	suite.router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mobile/devices", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	data := response["data"].([]interface{})
	assert.Len(suite.T(), data, 2)
}

func TestMobileHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MobileHandlerTestSuite))
}