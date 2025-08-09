//go:build unit

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/42atomys/webhooked/cmd/flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteWebhookedCmd struct {
	suite.Suite

	originalFlags struct {
		Config   string
		Help     bool
		Init     bool
		Port     int
		Validate bool
		Version  bool
		Debug    bool
	}
	tempConfigPath     string
	validConfigContent string
}

func (suite *TestSuiteWebhookedCmd) BeforeTest(suiteName, testName string) {
	// Save original flag values
	suite.originalFlags.Config = flags.Config
	suite.originalFlags.Help = flags.Help
	suite.originalFlags.Init = flags.Init
	suite.originalFlags.Port = flags.Port
	suite.originalFlags.Validate = flags.Validate
	suite.originalFlags.Version = flags.Version
	suite.originalFlags.Debug = flags.Debug

	// Setup temp config path
	suite.tempConfigPath = filepath.Join(os.TempDir(), "webhooked_test_config.yaml")
	suite.validConfigContent = `apiVersion: v1alpha2
kind: Configuration
metadata:
  name: test-config
specs:
- metricsEnabled: true
  webhooks:
  - name: test-webhook
    entrypointUrl: /test
    security:
      type: noop
    storage:
    - type: noop
`
}

func (suite *TestSuiteWebhookedCmd) AfterTest(suiteName, testName string) {
	// Restore original flag values
	flags.Config = suite.originalFlags.Config
	flags.Help = suite.originalFlags.Help
	flags.Init = suite.originalFlags.Init
	flags.Port = suite.originalFlags.Port
	flags.Validate = suite.originalFlags.Validate
	flags.Version = suite.originalFlags.Version
	flags.Debug = suite.originalFlags.Debug

	// Clean up temp files
	os.Remove(suite.tempConfigPath)
}

func (suite *TestSuiteWebhookedCmd) TestExec_Version() {
	assert := assert.New(suite.T())

	// Set version flag
	flags.Version = true
	flags.Config = "dummy.yaml" // Valid config path

	err := exec(context.Background())

	assert.NoError(err)
}

func (suite *TestSuiteWebhookedCmd) TestExec_Help() {
	assert := assert.New(suite.T())

	// Set help flag
	flags.Help = true
	flags.Config = "dummy.yaml" // Valid config path

	err := exec(context.Background())

	assert.NoError(err)
}

func (suite *TestSuiteWebhookedCmd) TestExec_Init() {
	assert := assert.New(suite.T())

	// Set init flag with non-existent config path
	flags.Init = true
	flags.Config = suite.tempConfigPath

	err := exec(context.Background())

	assert.NoError(err)
	// Check that config file was created
	_, err = os.Stat(suite.tempConfigPath)
	assert.NoError(err)
}

func (suite *TestSuiteWebhookedCmd) TestExec_InitExistingFile() {
	assert := assert.New(suite.T())

	// Create existing file
	err := os.WriteFile(suite.tempConfigPath, []byte("existing"), 0600)
	suite.Require().NoError(err)

	// Set init flag
	flags.Init = true
	flags.Config = suite.tempConfigPath

	err = exec(context.Background())

	assert.Error(err)
	assert.Contains(err.Error(), "configuration file already exists")
}

func (suite *TestSuiteWebhookedCmd) TestExec_Validate_ValidConfig() {
	assert := assert.New(suite.T())

	// Create valid config file
	err := os.WriteFile(suite.tempConfigPath, []byte(suite.validConfigContent), 0600)
	suite.Require().NoError(err)

	// Set validate flag
	flags.Validate = true
	flags.Config = suite.tempConfigPath

	err = exec(context.Background())

	assert.NoError(err)
}

func (suite *TestSuiteWebhookedCmd) TestExec_Validate_InvalidConfig() {
	assert := assert.New(suite.T())

	// Create invalid config file
	invalidConfig := "invalid: yaml: content ["
	err := os.WriteFile(suite.tempConfigPath, []byte(invalidConfig), 0600)
	suite.Require().NoError(err)

	// Set validate flag
	flags.Validate = true
	flags.Config = suite.tempConfigPath

	err = exec(context.Background())

	assert.Error(err)
	assert.Contains(err.Error(), "configuration validation failed")
}

func (suite *TestSuiteWebhookedCmd) TestExec_Validate_NonexistentConfig() {
	assert := assert.New(suite.T())

	// Set validate flag with non-existent config
	flags.Validate = true
	flags.Config = "/nonexistent/config.yaml"

	err := exec(context.Background())

	assert.Error(err)
	assert.Contains(err.Error(), "configuration validation failed")
}

func (suite *TestSuiteWebhookedCmd) TestExec_InvalidFlags() {
	assert := assert.New(suite.T())

	// Set invalid port
	flags.Port = 999999 // Invalid port
	flags.Config = "dummy.yaml"

	err := exec(context.Background())

	assert.Error(err)
	assert.Contains(err.Error(), "error validating flags")
}

func (suite *TestSuiteWebhookedCmd) TestExec_ConfigLoadError() {
	assert := assert.New(suite.T())

	// Use non-existent config file (not validation mode)
	flags.Config = "/nonexistent/config.yaml"
	flags.Port = 8080

	err := exec(context.Background())

	assert.Error(err)
	assert.Contains(err.Error(), "error loading config")
}

func (suite *TestSuiteWebhookedCmd) TestExec_ServerCreationError() {
	assert := assert.New(suite.T())

	// Create invalid config that will fail config loading (empty entrypoint URL)
	invalidServerConfig := `apiVersion: v1alpha2
kind: Configuration
specs:
- webhooks:
  - name: invalid-webhook
    entrypointUrl: ""
    security:
      type: noop
`
	err := os.WriteFile(suite.tempConfigPath, []byte(invalidServerConfig), 0600)
	suite.Require().NoError(err)

	flags.Config = suite.tempConfigPath
	flags.Port = 8080

	err = exec(context.Background())

	assert.Error(err)
	assert.Contains(err.Error(), "error loading config")
}

func (suite *TestSuiteWebhookedCmd) TestInitializeConfig_AbsolutePath() {
	assert := assert.New(suite.T())

	flags.Config = suite.tempConfigPath

	err := initializeConfig()

	assert.NoError(err)
	// Check that config file was created
	_, err = os.Stat(suite.tempConfigPath)
	assert.NoError(err)
}

func (suite *TestSuiteWebhookedCmd) TestInitializeConfig_RelativePath() {
	assert := assert.New(suite.T())

	// Use relative path
	relativePath := "test_webhooked_config.yaml"
	flags.Config = relativePath

	err := initializeConfig()

	assert.NoError(err)
	// Check that config file was created in current directory
	wd, _ := os.Getwd()
	fullPath := filepath.Join(wd, relativePath)
	_, err = os.Stat(fullPath)
	assert.NoError(err)

	// Clean up
	os.Remove(fullPath)
}

func (suite *TestSuiteWebhookedCmd) TestInitializeConfig_ExistingFile() {
	assert := assert.New(suite.T())

	// Create existing file
	err := os.WriteFile(suite.tempConfigPath, []byte("existing"), 0600)
	suite.Require().NoError(err)

	flags.Config = suite.tempConfigPath

	err = initializeConfig()

	assert.Error(err)
	assert.Contains(err.Error(), "configuration file already exists")
}

func (suite *TestSuiteWebhookedCmd) TestInitializeConfig_WriteError() {
	assert := assert.New(suite.T())

	// Use invalid path that will cause write error
	flags.Config = "/root/cannot_write_here.yaml"

	err := initializeConfig()

	assert.Error(err)
	assert.Contains(err.Error(), "failed to write configuration file")
}

func (suite *TestSuiteWebhookedCmd) TestApp_GracefulShutdown_NilServer() {
	assert := assert.New(suite.T())

	app := &app{server: nil}

	err := app.gracefulShutdown()

	assert.NoError(err)
}

// Note: Detailed server shutdown testing requires integration tests
// due to webhooked.Server type constraints

func (suite *TestSuiteWebhookedCmd) TestExec_DebugFlag() {
	assert := assert.New(suite.T())

	// Set debug flag and version flag (to exit early)
	flags.Debug = true
	flags.Version = true
	flags.Config = "dummy.yaml"

	err := exec(context.Background())

	assert.NoError(err)
	// Debug flag changes logging level, but we can't easily test that in unit tests
	// The important thing is that it doesn't cause errors
}

func TestRunWebhookedCmdSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteWebhookedCmd))
}

// Note: Mock server removed due to type constraints with *webhooked.Server

// Benchmarks
func BenchmarkInitializeConfig(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	// Save and restore flags
	originalConfig := flags.Config
	defer func() {
		flags.Config = originalConfig
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tempPath := filepath.Join(os.TempDir(), "bench_config.yaml")
		flags.Config = tempPath
		initializeConfig()  // nolint:errcheck
		os.Remove(tempPath) // Clean up
	}
}

func BenchmarkGracefulShutdown_NilServer(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	app := &app{server: nil}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.gracefulShutdown() // nolint:errcheck
	}
}
