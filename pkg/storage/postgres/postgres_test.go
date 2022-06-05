package postgres

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PostgresSetupTestSuite struct {
	suite.Suite
	client *sql.DB
}

// Create Table for running test
func (suite *PostgresSetupTestSuite) BeforeTest(suiteName, testName string) {
	var err error
	if suite.client, err = sql.Open("postgres", "postgresql://webhook:test@127.0.0.1:5432/webhook_db?sslmode=disable"); err != nil {
		suite.T().Error(err)
	}
	if _, err := suite.client.Query("CREATE TABLE test (test_field TEXT)"); err != nil {
		suite.T().Error(err)
	}
}

// Delete Table after test
func (suite *PostgresSetupTestSuite) AfterTest(suiteName, testName string) {
	if _, err := suite.client.Query("DROP TABLE test"); err != nil {
		suite.T().Error(err)
	}
}

func (suite *PostgresSetupTestSuite) TestPostgresName() {
	newPostgres := storage{}
	assert.Equal(suite.T(), "postgres", newPostgres.Name())
}

func (suite *PostgresSetupTestSuite) TestPostgresNewStorage() {
	_, err := NewStorage(map[string]interface{}{
		"databaseUrl": []int{1},
	})
	assert.Error(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"databaseUrl": "postgresql://webhook:test@127.0.0.1:5432/webhook_db?sslmode=disable",
		"tableName":   "test",
		"dataField":   "test_field",
	})
	assert.NoError(suite.T(), err)
}

func (suite *PostgresSetupTestSuite) TestPostgresPush() {
	newClient, _ := NewStorage(map[string]interface{}{
		"databaseUrl": "postgresql://webhook:test@127.0.0.1:5432/webhook_db?sslmode=disable",
		"tableName":   "Not Exist",
		"dataField":   "Not exist",
	})
	err := newClient.Push("Hello")
	assert.Error(suite.T(), err)

	newClient, err = NewStorage(map[string]interface{}{
		"databaseUrl": "postgresql://webhook:test@127.0.0.1:5432/webhook_db?sslmode=disable",
		"tableName":   "test",
		"dataField":   "test_field",
	})
	assert.NoError(suite.T(), err)

	err = newClient.Push("Hello")
	assert.NoError(suite.T(), err)
}

func TestRunPostgresPush(t *testing.T) {
	if testing.Short() {
		t.Skip("postgresql testing is skiped in short version of test")
		return
	}

	suite.Run(t, new(PostgresSetupTestSuite))
}
