//go:build unit

package postgres

import (
	"testing"

	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuitePostgresStorage struct {
	suite.Suite
}

func (suite *TestSuitePostgresStorage) TestEnsureConfigurationCompleteness_ValidConfig() {
	assert := assert.New(suite.T())

	databaseURL, err := valuable.Serialize("postgres://user:pass@localhost/db")
	require.NoError(suite.T(), err)

	spec := &PostgresStorageSpec{
		DatabaseURL: *databaseURL,
		Query:       "INSERT INTO webhooks (data) VALUES (:data)",
		Args:        map[string]string{"data": "{{ .Payload }}"},
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
}

func (suite *TestSuitePostgresStorage) TestEnsureConfigurationCompleteness_MissingDatabaseURL() {
	assert := assert.New(suite.T())

	emptyURL, err := valuable.Serialize("")
	require.NoError(suite.T(), err)

	spec := &PostgresStorageSpec{
		DatabaseURL: *emptyURL,
		Query:       "INSERT INTO webhooks (data) VALUES (:data)",
		Args:        map[string]string{"data": "{{ .Payload }}"},
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.Error(err)
	assert.Contains(err.Error(), "databaseUrl is required")
}

func (suite *TestSuitePostgresStorage) TestEnsureConfigurationCompleteness_MissingQuery() {
	assert := assert.New(suite.T())

	databaseURL, err := valuable.Serialize("postgres://user:pass@localhost/db")
	require.NoError(suite.T(), err)

	spec := &PostgresStorageSpec{
		DatabaseURL: *databaseURL,
		Query:       "",
		Args:        map[string]string{"data": "{{ .Payload }}"},
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.Error(err)
	assert.Contains(err.Error(), "query is required")
}

func (suite *TestSuitePostgresStorage) TestEnsureConfigurationCompleteness_NilArgs() {
	assert := assert.New(suite.T())

	databaseURL, err := valuable.Serialize("postgres://user:pass@localhost/db")
	require.NoError(suite.T(), err)

	spec := &PostgresStorageSpec{
		DatabaseURL: *databaseURL,
		Query:       "INSERT INTO webhooks (data) VALUES (:data)",
		Args:        nil,
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
	assert.NotNil(spec.Args) // Should be initialized
	assert.Empty(spec.Args)  // Should be empty map
}

func (suite *TestSuitePostgresStorage) TestEnsureConfigurationCompleteness_EmptyArgs() {
	assert := assert.New(suite.T())

	databaseURL, err := valuable.Serialize("postgres://user:pass@localhost/db")
	require.NoError(suite.T(), err)

	spec := &PostgresStorageSpec{
		DatabaseURL: *databaseURL,
		Query:       "INSERT INTO webhooks DEFAULT VALUES",
		Args:        map[string]string{},
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
}

func (suite *TestSuitePostgresStorage) TestEnsureConfigurationCompleteness_MultipleCalls() {
	assert := assert.New(suite.T())

	databaseURL, err := valuable.Serialize("postgres://user:pass@localhost/db")
	require.NoError(suite.T(), err)

	spec := &PostgresStorageSpec{
		DatabaseURL: *databaseURL,
		Query:       "INSERT INTO webhooks (data) VALUES (:data)",
		Args:        nil,
	}

	// First call should initialize Args
	err = spec.EnsureConfigurationCompleteness()
	assert.NoError(err)
	assert.NotNil(spec.Args)

	// Second call should not change anything
	originalArgs := spec.Args
	err = spec.EnsureConfigurationCompleteness()
	assert.NoError(err)
	assert.Equal(originalArgs, spec.Args)
}

func (suite *TestSuitePostgresStorage) TestStructInitialization() {
	assert := assert.New(suite.T())

	// Test that struct can be initialized in different ways
	spec1 := &PostgresStorageSpec{}
	spec2 := new(PostgresStorageSpec)
	var spec3 PostgresStorageSpec

	specs := []*PostgresStorageSpec{spec1, spec2, &spec3}

	for i, spec := range specs {
		assert.NotNil(spec, "Spec %d should not be nil", i)
		assert.Equal("", spec.Query, "Spec %d should have empty Query initially", i)
		assert.Nil(spec.Args, "Spec %d should have nil Args initially", i)
		assert.Nil(spec.client, "Spec %d should have nil client initially", i)
		assert.Nil(spec.formatters, "Spec %d should have nil formatters initially", i)
	}
}

func (suite *TestSuitePostgresStorage) TestFieldTypes() {
	assert := assert.New(suite.T())

	spec := &PostgresStorageSpec{}

	// Verify field types
	assert.IsType(valuable.Valuable{}, spec.DatabaseURL)
	assert.IsType("", spec.Query)
	assert.IsType(map[string]string(nil), spec.Args)
}

func TestRunPostgresStorageSuite(t *testing.T) {
	suite.Run(t, new(TestSuitePostgresStorage))
}

// Note: Initialize() and Store() methods require actual database connections
// and are better tested in integration tests. Unit tests focus on configuration
// validation and struct behavior.