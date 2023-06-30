package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"atomys.codes/webhooked/pkg/formatting"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PostgresSetupTestSuite struct {
	suite.Suite
	client      *sqlx.DB
	databaseUrl string
	ctx         context.Context
}

// Create Table for running test
func (suite *PostgresSetupTestSuite) BeforeTest(suiteName, testName string) {
	var err error

	suite.databaseUrl = fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	if suite.client, err = sqlx.Open("postgres", suite.databaseUrl); err != nil {
		suite.T().Error(err)
	}
	if _, err := suite.client.Query("CREATE TABLE test (test_field TEXT)"); err != nil {
		suite.T().Error(err)
	}

	suite.ctx = formatting.ToContext(
		context.Background(),
		formatting.New().WithTemplate("{{.}}"),
	)

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
		"databaseUrl": suite.databaseUrl,
		"tableName":   "test",
		"dataField":   "test_field",
	})
	assert.NoError(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"databaseUrl":                 suite.databaseUrl,
		"tableName":                   "test",
		"useFormattingToPerformQuery": true,
	})
	assert.Error(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"databaseUrl":                 suite.databaseUrl,
		"useFormattingToPerformQuery": true,
		"query":                       "",
	})
	assert.Error(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"databaseUrl":                 suite.databaseUrl,
		"useFormattingToPerformQuery": true,
		"query":                       "INSERT INTO test (test_field) VALUES ('$field')",
	})
	assert.NoError(suite.T(), err)
}

func (suite *PostgresSetupTestSuite) TestPostgresPush() {
	newClient, _ := NewStorage(map[string]interface{}{
		"databaseUrl": suite.databaseUrl,
		"tableName":   "Not Exist",
		"dataField":   "Not exist",
	})
	err := newClient.Push(suite.ctx, []byte("Hello"))
	assert.Error(suite.T(), err)

	newClient, err = NewStorage(map[string]interface{}{
		"databaseUrl": suite.databaseUrl,
		"tableName":   "test",
		"dataField":   "test_field",
	})
	assert.NoError(suite.T(), err)

	err = newClient.Push(suite.ctx, []byte("Hello"))
	assert.NoError(suite.T(), err)
}

func (suite *PostgresSetupTestSuite) TestPostgresPushNewFormattedQuery() {
	newClient, err := NewStorage(map[string]interface{}{
		"databaseUrl":                 suite.databaseUrl,
		"useFormattingToPerformQuery": true,
		"query":                       "INSERT INTO test (test_field) VALUES (:field)",
		"args": map[string]string{
			"field": "{{.Payload}}",
		},
	})
	assert.NoError(suite.T(), err)

	fakePayload := []byte("A strange payload")
	err = newClient.Push(
		suite.ctx,
		fakePayload,
	)
	assert.NoError(suite.T(), err)

	rows, err := suite.client.Query("SELECT test_field FROM test")
	assert.NoError(suite.T(), err)

	var result string
	for rows.Next() {
		err := rows.Scan(&result)
		assert.NoError(suite.T(), err)
	}
	assert.Equal(suite.T(), string(fakePayload), result)
}

func TestRunPostgresPush(t *testing.T) {
	if testing.Short() {
		t.Skip("postgresql testing is skiped in short version of test")
		return
	}

	suite.Run(t, new(PostgresSetupTestSuite))
}
