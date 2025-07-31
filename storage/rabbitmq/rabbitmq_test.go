//go:build unit

package rabbitmq

import (
	"testing"

	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuiteRabbitmqStorage struct {
	suite.Suite
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_DefaultValues() {
	assert := assert.New(suite.T())

	spec := &RabbitmqStorageSpec{}

	err := spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
	assert.Equal("text/plain", spec.DefinedContentType)
	assert.Equal(5, spec.MaxAttempt)
	assert.NotNil(spec.Durable)
	assert.True(*spec.Durable)
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_ExistingValues() {
	assert := assert.New(suite.T())

	durable := false
	spec := &RabbitmqStorageSpec{
		DefinedContentType: "application/json",
		MaxAttempt:         10,
		Durable:            &durable,
	}

	err := spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
	// Should not overwrite existing values
	assert.Equal("application/json", spec.DefinedContentType)
	assert.Equal(10, spec.MaxAttempt)
	assert.NotNil(spec.Durable)
	assert.False(*spec.Durable)
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_EmptyContentType() {
	assert := assert.New(suite.T())

	spec := &RabbitmqStorageSpec{
		DefinedContentType: "",
		MaxAttempt:         3,
	}

	err := spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
	assert.Equal("text/plain", spec.DefinedContentType) // Should set default
	assert.Equal(3, spec.MaxAttempt)                     // Should keep existing
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_ZeroMaxAttempt() {
	assert := assert.New(suite.T())

	spec := &RabbitmqStorageSpec{
		DefinedContentType: "application/xml",
		MaxAttempt:         0,
	}

	err := spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
	assert.Equal("application/xml", spec.DefinedContentType) // Should keep existing
	assert.Equal(5, spec.MaxAttempt)                         // Should set default
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_NilDurable() {
	assert := assert.New(suite.T())

	spec := &RabbitmqStorageSpec{
		Durable: nil,
	}

	err := spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
	assert.NotNil(spec.Durable)
	assert.True(*spec.Durable) // Should set default to true
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_MultipleCalls() {
	assert := assert.New(suite.T())

	spec := &RabbitmqStorageSpec{}

	// First call should set defaults
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)
	assert.Equal("text/plain", spec.DefinedContentType)
	assert.Equal(5, spec.MaxAttempt)
	assert.True(*spec.Durable)

	// Second call should not change anything
	err = spec.EnsureConfigurationCompleteness()
	assert.NoError(err)
	assert.Equal("text/plain", spec.DefinedContentType)
	assert.Equal(5, spec.MaxAttempt)
	assert.True(*spec.Durable)
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_DifferentContentTypes() {
	assert := assert.New(suite.T())

	contentTypes := []string{
		"application/json",
		"application/xml",
		"text/plain",
		"application/octet-stream",
		"text/html",
	}

	for _, contentType := range contentTypes {
		spec := &RabbitmqStorageSpec{
			DefinedContentType: contentType,
		}

		err := spec.EnsureConfigurationCompleteness()
		assert.NoError(err, "Content type %s should be valid", contentType)
		assert.Equal(contentType, spec.DefinedContentType)
	}
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_DifferentMaxAttempts() {
	assert := assert.New(suite.T())

	maxAttempts := []int{1, 3, 5, 10, 100}

	for _, maxAttempt := range maxAttempts {
		spec := &RabbitmqStorageSpec{
			MaxAttempt: maxAttempt,
		}

		err := spec.EnsureConfigurationCompleteness()
		assert.NoError(err, "MaxAttempt %d should be valid", maxAttempt)
		assert.Equal(maxAttempt, spec.MaxAttempt)
	}
}

func (suite *TestSuiteRabbitmqStorage) TestStructInitialization() {
	assert := assert.New(suite.T())

	// Test that struct can be initialized in different ways
	spec1 := &RabbitmqStorageSpec{}
	spec2 := new(RabbitmqStorageSpec)
	var spec3 RabbitmqStorageSpec

	specs := []*RabbitmqStorageSpec{spec1, spec2, &spec3}

	for i, spec := range specs {
		assert.NotNil(spec, "Spec %d should not be nil", i)
		assert.Equal("", spec.DefinedContentType, "Spec %d should have empty DefinedContentType initially", i)
		assert.Equal(0, spec.MaxAttempt, "Spec %d should have MaxAttempt 0 initially", i)
		assert.Equal("", spec.QueueName, "Spec %d should have empty QueueName initially", i)
		assert.Nil(spec.Durable, "Spec %d should have nil Durable initially", i)
		assert.False(spec.DeleteWhenUnused, "Spec %d should have DeleteWhenUnused false initially", i)
		assert.False(spec.Exclusive, "Spec %d should have Exclusive false initially", i)
		assert.False(spec.NoWait, "Spec %d should have NoWait false initially", i)
		assert.Equal("", spec.Exchange, "Spec %d should have empty Exchange initially", i)
		assert.False(spec.Mandatory, "Spec %d should have Mandatory false initially", i)
		assert.False(spec.Immediate, "Spec %d should have Immediate false initially", i)
		assert.Nil(spec.client, "Spec %d should have nil client initially", i)
		assert.Nil(spec.channel, "Spec %d should have nil channel initially", i)
	}
}

func (suite *TestSuiteRabbitmqStorage) TestFieldTypes() {
	assert := assert.New(suite.T())

	spec := &RabbitmqStorageSpec{}

	// Verify field types
	assert.IsType(valuable.Valuable{}, spec.DatabaseURL)
	assert.IsType(0, spec.MaxAttempt)
	assert.IsType("", spec.QueueName)
	assert.IsType((*bool)(nil), spec.Durable)
	assert.IsType(false, spec.DeleteWhenUnused)
	assert.IsType(false, spec.Exclusive)
	assert.IsType(false, spec.NoWait)
	assert.IsType("", spec.Exchange)
	assert.IsType("", spec.DefinedContentType)
	assert.IsType(false, spec.Mandatory)
	assert.IsType(false, spec.Immediate)
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_BooleanFlags() {
	assert := assert.New(suite.T())

	spec := &RabbitmqStorageSpec{
		DeleteWhenUnused: true,
		Exclusive:        true,
		NoWait:           true,
		Mandatory:        true,
		Immediate:        true,
	}

	err := spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
	// Boolean flags should remain unchanged
	assert.True(spec.DeleteWhenUnused)
	assert.True(spec.Exclusive)
	assert.True(spec.NoWait)
	assert.True(spec.Mandatory)
	assert.True(spec.Immediate)
}

func (suite *TestSuiteRabbitmqStorage) TestEnsureConfigurationCompleteness_CompleteConfig() {
	assert := assert.New(suite.T())

	databaseURL, err := valuable.Serialize("amqp://guest:guest@localhost:5672/")
	require.NoError(suite.T(), err)

	durable := false
	spec := &RabbitmqStorageSpec{
		DatabaseURL:        *databaseURL,
		MaxAttempt:         3,
		QueueName:          "webhooks",
		Durable:            &durable,
		DeleteWhenUnused:   false,
		Exclusive:          false,
		NoWait:             false,
		Exchange:           "webhook-exchange",
		DefinedContentType: "application/json",
		Mandatory:          true,
		Immediate:          false,
	}

	err = spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
	// All values should remain unchanged
	assert.Equal(3, spec.MaxAttempt)
	assert.Equal("webhooks", spec.QueueName)
	assert.False(*spec.Durable)
	assert.False(spec.DeleteWhenUnused)
	assert.False(spec.Exclusive)
	assert.False(spec.NoWait)
	assert.Equal("webhook-exchange", spec.Exchange)
	assert.Equal("application/json", spec.DefinedContentType)
	assert.True(spec.Mandatory)
	assert.False(spec.Immediate)
}

func TestRunRabbitmqStorageSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteRabbitmqStorage))
}

// Note: Initialize() and Store() methods require actual RabbitMQ connections
// and are better tested in integration tests. Unit tests focus on configuration
// validation and struct behavior.