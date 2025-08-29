//go:build unit

package redis

import (
	"testing"

	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuiteRedisStorage struct {
	suite.Suite
}

func (suite *TestSuiteRedisStorage) TestEnsureConfigurationCompleteness_ValidConfig() {
	assert := assert.New(suite.T())

	host, err := valuable.Serialize("localhost")
	require.NoError(suite.T(), err)
	
	port, err := valuable.Serialize("6379")
	require.NoError(suite.T(), err)

	username, err := valuable.Serialize("user")
	require.NoError(suite.T(), err)

	password, err := valuable.Serialize("pass")
	require.NoError(suite.T(), err)

	spec := &RedisStorageSpec{
		Host:     *host,
		Port:     *port,
		Username: *username,
		Password: *password,
		Database: 0,
		Key:      "webhooks",
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
}

func (suite *TestSuiteRedisStorage) TestEnsureConfigurationCompleteness_MissingHost() {
	assert := assert.New(suite.T())

	emptyHost, err := valuable.Serialize("")
	require.NoError(suite.T(), err)
	
	port, err := valuable.Serialize("6379")
	require.NoError(suite.T(), err)

	spec := &RedisStorageSpec{
		Host: *emptyHost,
		Port: *port,
		Key:  "webhooks",
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.Error(err)
	assert.Contains(err.Error(), "host is required")
}

func (suite *TestSuiteRedisStorage) TestEnsureConfigurationCompleteness_MissingPort() {
	assert := assert.New(suite.T())

	host, err := valuable.Serialize("localhost")
	require.NoError(suite.T(), err)
	
	emptyPort, err := valuable.Serialize("")
	require.NoError(suite.T(), err)

	spec := &RedisStorageSpec{
		Host: *host,
		Port: *emptyPort,
		Key:  "webhooks",
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.Error(err)
	assert.Contains(err.Error(), "port is required")
}

func (suite *TestSuiteRedisStorage) TestEnsureConfigurationCompleteness_BothMissing() {
	assert := assert.New(suite.T())

	emptyHost, err := valuable.Serialize("")
	require.NoError(suite.T(), err)
	
	emptyPort, err := valuable.Serialize("")
	require.NoError(suite.T(), err)

	spec := &RedisStorageSpec{
		Host: *emptyHost,
		Port: *emptyPort,
		Key:  "webhooks",
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.Error(err)
	assert.Contains(err.Error(), "host is required")
}

func (suite *TestSuiteRedisStorage) TestEnsureConfigurationCompleteness_OptionalFields() {
	assert := assert.New(suite.T())

	host, err := valuable.Serialize("localhost")
	require.NoError(suite.T(), err)
	
	port, err := valuable.Serialize("6379")
	require.NoError(suite.T(), err)

	// Test with empty optional fields
	emptyUsername, err := valuable.Serialize("")
	require.NoError(suite.T(), err)

	emptyPassword, err := valuable.Serialize("")
	require.NoError(suite.T(), err)

	spec := &RedisStorageSpec{
		Host:     *host,
		Port:     *port,
		Username: *emptyUsername,
		Password: *emptyPassword,
		Database: 0,
		Key:      "",
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
}

func (suite *TestSuiteRedisStorage) TestEnsureConfigurationCompleteness_DifferentDatabases() {
	assert := assert.New(suite.T())

	host, err := valuable.Serialize("localhost")
	require.NoError(suite.T(), err)
	
	port, err := valuable.Serialize("6379")
	require.NoError(suite.T(), err)

	databases := []int{0, 1, 5, 15}

	for _, db := range databases {
		spec := &RedisStorageSpec{
			Host:     *host,
			Port:     *port,
			Database: db,
			Key:      "webhooks",
		}

		err = spec.EnsureConfigurationCompleteness()
		assert.NoError(err, "Database %d should be valid", db)
	}
}

func (suite *TestSuiteRedisStorage) TestEnsureConfigurationCompleteness_DifferentPorts() {
	assert := assert.New(suite.T())

	host, err := valuable.Serialize("localhost")
	require.NoError(suite.T(), err)

	ports := []string{"6379", "6380", "16379", "26379"}

	for _, portStr := range ports {
		port, err := valuable.Serialize(portStr)
		require.NoError(suite.T(), err)

		spec := &RedisStorageSpec{
			Host: *host,
			Port: *port,
			Key:  "webhooks",
		}

		err = spec.EnsureConfigurationCompleteness()
		assert.NoError(err, "Port %s should be valid", portStr)
	}
}

func (suite *TestSuiteRedisStorage) TestStructInitialization() {
	assert := assert.New(suite.T())

	// Test that struct can be initialized in different ways
	spec1 := &RedisStorageSpec{}
	spec2 := new(RedisStorageSpec)
	var spec3 RedisStorageSpec

	specs := []*RedisStorageSpec{spec1, spec2, &spec3}

	for i, spec := range specs {
		assert.NotNil(spec, "Spec %d should not be nil", i)
		assert.Equal(0, spec.Database, "Spec %d should have Database 0 initially", i)
		assert.Equal("", spec.Key, "Spec %d should have empty Key initially", i)
		assert.Nil(spec.client, "Spec %d should have nil client initially", i)
	}
}

func (suite *TestSuiteRedisStorage) TestFieldTypes() {
	assert := assert.New(suite.T())

	spec := &RedisStorageSpec{}

	// Verify field types
	assert.IsType(valuable.Valuable{}, spec.Host)
	assert.IsType(valuable.Valuable{}, spec.Port)
	assert.IsType(valuable.Valuable{}, spec.Username)
	assert.IsType(valuable.Valuable{}, spec.Password)
	assert.IsType(0, spec.Database)
	assert.IsType("", spec.Key)
}

func (suite *TestSuiteRedisStorage) TestEnsureConfigurationCompleteness_MultipleCalls() {
	assert := assert.New(suite.T())

	host, err := valuable.Serialize("localhost")
	require.NoError(suite.T(), err)
	
	port, err := valuable.Serialize("6379")
	require.NoError(suite.T(), err)

	spec := &RedisStorageSpec{
		Host: *host,
		Port: *port,
		Key:  "webhooks",
	}

	// Multiple calls should be idempotent
	err = spec.EnsureConfigurationCompleteness()
	assert.NoError(err)

	err = spec.EnsureConfigurationCompleteness()
	assert.NoError(err)
}

func TestRunRedisStorageSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteRedisStorage))
}

// Note: Initialize() and Store() methods require actual Redis connections
// and are better tested in integration tests. Unit tests focus on configuration
// validation and struct behavior.