package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/labmino/runsight-backend/internal/handlers"
	"github.com/labmino/runsight-backend/internal/middleware"
	"github.com/labmino/runsight-backend/tests/testhelpers"
)

type APITestSuite struct {
	suite.Suite
	router  *gin.Engine
	db      interface{}
	cleanup func()
}

func (suite *APITestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	db, cleanup := testhelpers.SetupTestDB()
	suite.db = db
	suite.cleanup = cleanup

	suite.router = gin.New()
	
	authHandler := handlers.NewAuthHandler(db)
	mobileHandler := handlers.NewMobileHandler(db)

	api := suite.router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "RunSight API is running",
			})
		})

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protectedAuth := api.Group("/auth")
		protectedAuth.Use(middleware.AuthMiddleware())
		{
			protectedAuth.GET("/profile", authHandler.GetProfile)
			protectedAuth.PUT("/profile", authHandler.UpdateProfile)
		}

		mobile := api.Group("/mobile")
		mobile.Use(middleware.AuthMiddleware())
		{
			mobile.POST("/pairing/request", mobileHandler.RequestPairingCode)
			mobile.GET("/pairing/:session_id/status", mobileHandler.CheckPairingStatus)
			mobile.GET("/devices", mobileHandler.GetDevices)
			mobile.DELETE("/devices/:device_id", mobileHandler.RemoveDevice)
			mobile.GET("/runs", mobileHandler.ListRuns)
			mobile.GET("/runs/:run_id", mobileHandler.GetRun)
			mobile.PATCH("/runs/:run_id", mobileHandler.UpdateRunNotes)
			mobile.GET("/stats", mobileHandler.GetStats)
		}
	}
}

func (suite *APITestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *APITestSuite) TestHealthEndpoint() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ok", response["status"])
	assert.Equal(suite.T(), "RunSight API is running", response["message"])
}

func (suite *APITestSuite) TestUserRegistration() {
	userData := map[string]interface{}{
		"full_name":        "John Doe",
		"email":            "john@example.com",
		"phone":            "+1234567890",
		"password":         "password123",
		"confirm_password": "password123",
	}
	jsonData, _ := json.Marshal(userData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(suite.T(), data["token"])
	
	user := data["user"].(map[string]interface{})
	assert.Equal(suite.T(), "John Doe", user["full_name"])
	assert.Equal(suite.T(), "john@example.com", user["email"])
}

func (suite *APITestSuite) TestUserLogin() {
	userData := map[string]interface{}{
		"full_name":        "Jane Doe",
		"email":            "jane@example.com",
		"phone":            "+1234567890",
		"password":         "password123",
		"confirm_password": "password123",
	}
	jsonData, _ := json.Marshal(userData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	loginData := map[string]interface{}{
		"email":    "jane@example.com",
		"password": "password123",
	}
	loginJSON, _ := json.Marshal(loginData)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginJSON))
	req2.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response["status"])
	
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(suite.T(), data["token"])
}

func (suite *APITestSuite) TestInvalidRegistration() {
	userData := map[string]interface{}{
		"full_name": "",
		"email":     "invalid-email",
		"password":  "123",
	}
	jsonData, _ := json.Marshal(userData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *APITestSuite) TestUnauthorizedEndpoints() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/mobile/runs", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}