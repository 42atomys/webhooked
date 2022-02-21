package storages

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
	if suite.client, err = sql.Open("postgres", "postgresql://webhook:test@localhost:5432/webhook_db?sslmode=disable"); err != nil {
		assert.Error(suite.T(), err)
	}
	if _, err := suite.client.Query("CREATE TABLE test (test_field TEXT)"); err != nil {
		assert.Error(suite.T(), err)
	}
}

// Delete Table after test
func (suite *PostgresSetupTestSuite) AfterTest(suiteName, testName string) {
	if _, err := suite.client.Query("DROP TABLE test"); err != nil {
		assert.Error(suite.T(), err)
	}
}

func TestPostgresName(t *testing.T) {
	newPostgres := PostgresStorage{}
	assert.Equal(t, "postgres", newPostgres.Name())
}

func TestPostgresNewPostgresStorage(t *testing.T) {
	storageSpec := map[string]interface{}{
		"databaseURL": "postgresql://webhook:test@localhost:5432/webhook_db?sslmode=disable",
		"tableName":   "test",
		"dataField":   "test_field",
	}

	_, err := NewPostgresStorage(storageSpec)
	assert.Nil(t, err)
}

func (suite *PostgresSetupTestSuite) TestPostgresPush() {
	newClient, err := NewPostgresStorage(map[string]interface{}{
		"databaseURL": "postgresql://webhook:test@localhost:5432/webhook_db?sslmode=disable",
		"tableName":   "test",
		"dataField":   "test_field",
	})
	assert.Nil(suite.T(), err)

	err = newClient.Push("Hello")
	assert.Nil(suite.T(), err)
}
