package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/labmino/runsight-backend/tests/testhelpers"
)

type UserModelTestSuite struct {
	suite.Suite
	cleanup func()
}

func (suite *UserModelTestSuite) SetupSuite() {
	_, cleanup := testhelpers.SetupTestDB()
	suite.cleanup = cleanup
}

func (suite *UserModelTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *UserModelTestSuite) TestCreateUser() {
	db, cleanup := testhelpers.SetupTestDB()
	defer cleanup()

	user := testhelpers.CreateTestUser(db, "test@example.com")
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), "test@example.com", user.Email)
	assert.Equal(suite.T(), "Test User", user.FullName)
}

func TestUserModelTestSuite(t *testing.T) {
	suite.Run(t, new(UserModelTestSuite))
}