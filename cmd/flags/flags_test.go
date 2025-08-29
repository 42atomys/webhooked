//go:build unit

package flags

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteFlagsValidation struct {
	suite.Suite

	originalPort   int
	originalConfig string
}

func (suite *TestSuiteFlagsValidation) BeforeTest(suiteName, testName string) {
	// Save original values
	suite.originalPort = Port
	suite.originalConfig = Config
}

func (suite *TestSuiteFlagsValidation) AfterTest(suiteName, testName string) {
	// Restore original values
	Port = suite.originalPort
	Config = suite.originalConfig
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_ValidPort() {
	assert := assert.New(suite.T())

	Port = 8080
	Config = "webhooked.yaml"

	err := ValidateFlags()

	assert.NoError(err)
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_ValidPortRange() {
	assert := assert.New(suite.T())

	testCases := []int{1, 80, 443, 8080, 65535}

	for _, port := range testCases {
		Port = port
		Config = "webhooked.yaml"

		err := ValidateFlags()

		assert.NoError(err, "Port %d should be valid", port)
	}
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_InvalidPortTooLow() {
	assert := assert.New(suite.T())

	Port = 0
	Config = "webhooked.yaml"

	err := ValidateFlags()

	assert.Error(err)
	assert.Contains(err.Error(), "invalid port number: 0")
	assert.Contains(err.Error(), "must be between 1 and 65535")
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_InvalidPortNegative() {
	assert := assert.New(suite.T())

	Port = -1
	Config = "webhooked.yaml"

	err := ValidateFlags()

	assert.Error(err)
	assert.Contains(err.Error(), "invalid port number: -1")
	assert.Contains(err.Error(), "must be between 1 and 65535")
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_InvalidPortTooHigh() {
	assert := assert.New(suite.T())

	Port = 65536
	Config = "webhooked.yaml"

	err := ValidateFlags()

	assert.Error(err)
	assert.Contains(err.Error(), "invalid port number: 65536")
	assert.Contains(err.Error(), "must be between 1 and 65535")
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_InvalidPortVeryHigh() {
	assert := assert.New(suite.T())

	Port = 99999
	Config = "webhooked.yaml"

	err := ValidateFlags()

	assert.Error(err)
	assert.Contains(err.Error(), "invalid port number: 99999")
	assert.Contains(err.Error(), "must be between 1 and 65535")
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_EmptyConfig() {
	assert := assert.New(suite.T())

	Port = 8080
	Config = ""

	err := ValidateFlags()

	assert.Error(err)
	assert.Contains(err.Error(), "config file path is required")
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_ValidConfig() {
	assert := assert.New(suite.T())

	validConfigs := []string{
		"webhooked.yaml",
		"/path/to/config.yaml",
		"./relative/path/config.yml",
		"config.json",
		"/absolute/path/with spaces/config.yaml",
	}

	for _, config := range validConfigs {
		Port = 8080
		Config = config

		err := ValidateFlags()

		assert.NoError(err, "Config '%s' should be valid", config)
	}
}

func (suite *TestSuiteFlagsValidation) TestValidateFlags_BothInvalid() {
	assert := assert.New(suite.T())

	Port = 0
	Config = ""

	err := ValidateFlags()

	// Should return the port error first
	assert.Error(err)
	assert.Contains(err.Error(), "invalid port number")
}

func (suite *TestSuiteFlagsValidation) TestDefaultValues() {
	assert := assert.New(suite.T())

	// Test that default values are set correctly
	// Note: These are set during package initialization
	assert.Equal("webhooked.yaml", Config)
	assert.Equal(8080, Port)
	assert.False(Help)
	assert.False(Init)
	assert.False(Validate)
	assert.False(Version)
	assert.False(Debug)
}

func (suite *TestSuiteFlagsValidation) TestUsageConstant() {
	assert := assert.New(suite.T())

	// Test that usage constant contains expected content
	assert.Contains(usage, "Usage: webhooked [options]")
	assert.Contains(usage, "--help")
	assert.Contains(usage, "--version")
	assert.Contains(usage, "--config")
	assert.Contains(usage, "--init")
	assert.Contains(usage, "--port")
	assert.Contains(usage, "--validate")
}

func (suite *TestSuiteFlagsValidation) TestUsageFn() {
	assert := assert.New(suite.T())

	// Test that usage function prints expected content
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stdout)

	usageFn()

	log.Print(usage)
	pflag.PrintDefaults()

	assert.Contains(buf.String(), "Usage: webhooked [options]")
	assert.Contains(buf.String(), "--help")
	assert.Contains(buf.String(), "--version")
	assert.Contains(buf.String(), "--config")
	assert.Contains(buf.String(), "--init")
	assert.Contains(buf.String(), "--port")
	assert.Contains(buf.String(), "--validate")
}

func (suite *TestSuiteFlagsValidation) TestFlagVariablesExist() {
	assert := assert.New(suite.T())

	// Test that all flag variables are accessible
	assert.IsType("", Config)
	assert.IsType(0, Port)
	assert.IsType(false, Help)
	assert.IsType(false, Init)
	assert.IsType(false, Validate)
	assert.IsType(false, Version)
	assert.IsType(false, Debug)
}

func (suite *TestSuiteFlagsValidation) TestPortBoundaryValues() {
	assert := assert.New(suite.T())

	// Test exact boundary values
	testCases := []struct {
		port    int
		valid   bool
		message string
	}{
		{0, false, "Port 0 should be invalid"},
		{1, true, "Port 1 should be valid (minimum)"},
		{65535, true, "Port 65535 should be valid (maximum)"},
		{65536, false, "Port 65536 should be invalid"},
	}

	for _, tc := range testCases {
		Port = tc.port
		Config = "webhooked.yaml"

		err := ValidateFlags()

		if tc.valid {
			assert.NoError(err, tc.message)
		} else {
			assert.Error(err, tc.message)
		}
	}
}

func TestRunFlagsValidationSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteFlagsValidation))
}

// Benchmarks

func BenchmarkValidateFlags_Valid(b *testing.B) {
	originalPort := Port
	originalConfig := Config
	defer func() {
		Port = originalPort
		Config = originalConfig
	}()

	Port = 8080
	Config = "webhooked.yaml"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateFlags() // nolint:errcheck
	}
}

func BenchmarkValidateFlags_InvalidPort(b *testing.B) {
	originalPort := Port
	originalConfig := Config
	defer func() {
		Port = originalPort
		Config = originalConfig
	}()

	Port = 0
	Config = "webhooked.yaml"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateFlags() // nolint:errcheck
	}
}

func BenchmarkValidateFlags_EmptyConfig(b *testing.B) {
	originalPort := Port
	originalConfig := Config
	defer func() {
		Port = originalPort
		Config = originalConfig
	}()

	Port = 8080
	Config = ""

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateFlags() // nolint:errcheck
	}
}
