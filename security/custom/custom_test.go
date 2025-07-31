//go:build unit

package custom

import (
	"context"
	"testing"

	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type TestSuiteCustomSecurity struct {
	suite.Suite

	validCondition   *valuable.Valuable
	invalidCondition *valuable.Valuable
	emptyCondition   *valuable.Valuable
	trueCondition    *valuable.Valuable
	falseCondition   *valuable.Valuable
	ctx              context.Context
	requestCtx       *fasthttpz.RequestCtx
}

func (suite *TestSuiteCustomSecurity) BeforeTest(suiteName, testName string) {
	var err error

	// Valid condition that evaluates to true
	suite.validCondition, err = valuable.Serialize("true")
	require.NoError(suite.T(), err)

	// Invalid condition with syntax error
	suite.invalidCondition, err = valuable.Serialize("{{ invalid template")
	require.NoError(suite.T(), err)

	// Empty condition
	suite.emptyCondition, err = valuable.Serialize("")
	require.NoError(suite.T(), err)

	// Explicit true condition
	suite.trueCondition, err = valuable.Serialize("true")
	require.NoError(suite.T(), err)

	// Explicit false condition
	suite.falseCondition, err = valuable.Serialize("false")
	require.NoError(suite.T(), err)

	// Setup context and request context
	suite.ctx = context.Background()
	suite.requestCtx = &fasthttpz.RequestCtx{
		RequestCtx: &fasthttp.RequestCtx{},
	}
}

func (suite *TestSuiteCustomSecurity) TestEnsureConfigurationCompleteness_ValidCondition() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.validCondition,
	}

	err := spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
}

func (suite *TestSuiteCustomSecurity) TestEnsureConfigurationCompleteness_NilCondition() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: nil,
	}

	err := spec.EnsureConfigurationCompleteness()

	assert.Error(err)
	assert.Contains(err.Error(), "condition is required")
}

func (suite *TestSuiteCustomSecurity) TestEnsureConfigurationCompleteness_EmptyCondition() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.emptyCondition,
	}

	err := spec.EnsureConfigurationCompleteness()

	assert.Error(err)
	assert.Contains(err.Error(), "condition is required")
}

func (suite *TestSuiteCustomSecurity) TestInitialize_ValidCondition() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.validCondition,
	}

	err := spec.Initialize()

	assert.NoError(err)
	assert.NotNil(spec.formatter)
	assert.True(spec.formatter.HasTemplate())
	assert.True(spec.formatter.HasTemplateCompiled())
}

func (suite *TestSuiteCustomSecurity) TestInitialize_InvalidCondition() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.invalidCondition,
	}

	err := spec.Initialize()

	assert.Error(err)
	assert.Contains(err.Error(), "error compiling template")
}

func (suite *TestSuiteCustomSecurity) TestInitialize_EmptyCondition() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.emptyCondition,
	}

	err := spec.Initialize()

	assert.Error(err)
	assert.Contains(err.Error(), "condition template is required")
}

func (suite *TestSuiteCustomSecurity) TestIsSecure_TrueCondition() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.trueCondition,
	}
	err := spec.Initialize()
	require.NoError(suite.T(), err)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteCustomSecurity) TestIsSecure_FalseCondition() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.falseCondition,
	}
	err := spec.Initialize()
	require.NoError(suite.T(), err)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.False(result)
}

func (suite *TestSuiteCustomSecurity) TestIsSecure_TemplateExecutionError() {
	assert := assert.New(suite.T())

	// Create a condition that will cause template execution error
	errorCondition, err := valuable.Serialize("{{ .NonExistentField.SubField }}")
	require.NoError(suite.T(), err)

	spec := &CustomSecuritySpec{
		Condition: errorCondition,
	}
	err = spec.Initialize()
	require.NoError(suite.T(), err)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.Error(err)
	assert.Contains(err.Error(), "failed to parse custom security condition result as boolean")
	assert.False(result)
}

func (suite *TestSuiteCustomSecurity) TestIsSecure_InvalidBooleanResult() {
	assert := assert.New(suite.T())

	// Create a condition that evaluates to non-boolean value
	nonBoolCondition, err := valuable.Serialize("not_a_boolean")
	require.NoError(suite.T(), err)

	spec := &CustomSecuritySpec{
		Condition: nonBoolCondition,
	}
	err = spec.Initialize()
	require.NoError(suite.T(), err)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.Error(err)
	assert.Contains(err.Error(), "failed to parse custom security condition result as boolean")
	assert.False(result)
}

func (suite *TestSuiteCustomSecurity) TestIsSecure_TrueWithWhitespace() {
	assert := assert.New(suite.T())

	// Create a condition that evaluates to "true" with whitespace
	trueWithWhitespace, err := valuable.Serialize("true\n")
	require.NoError(suite.T(), err)

	spec := &CustomSecuritySpec{
		Condition: trueWithWhitespace,
	}
	err = spec.Initialize()
	require.NoError(suite.T(), err)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteCustomSecurity) TestIsSecure_FalseWithWhitespace() {
	assert := assert.New(suite.T())

	// Create a condition that evaluates to "false" with whitespace
	falseWithWhitespace, err := valuable.Serialize("false\n")
	require.NoError(suite.T(), err)

	spec := &CustomSecuritySpec{
		Condition: falseWithWhitespace,
	}
	err = spec.Initialize()
	require.NoError(suite.T(), err)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.False(result)
}

func (suite *TestSuiteCustomSecurity) TestIsSecure_ComplexCondition() {
	assert := assert.New(suite.T())

	// Create a more complex condition using template logic
	complexCondition, err := valuable.Serialize("{{ if eq 1 1 }}true{{ else }}false{{ end }}")
	require.NoError(suite.T(), err)

	spec := &CustomSecuritySpec{
		Condition: complexCondition,
	}
	err = spec.Initialize()
	require.NoError(suite.T(), err)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteCustomSecurity) TestIsSecure_UninitializedFormatter() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.validCondition,
		formatter: nil, // Not initialized
	}

	// This should panic when calling Format on nil formatter
	assert.Panics(func() {
		spec.IsSecure(suite.ctx, suite.requestCtx)
	})
}

func (suite *TestSuiteCustomSecurity) TestFullWorkflow_ValidConfiguration() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: suite.trueCondition,
	}

	// Test complete workflow
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)

	err = spec.Initialize()
	assert.NoError(err)

	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)
	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteCustomSecurity) TestFullWorkflow_InvalidConfiguration() {
	assert := assert.New(suite.T())

	spec := &CustomSecuritySpec{
		Condition: nil,
	}

	// Test workflow with invalid configuration
	err := spec.EnsureConfigurationCompleteness()
	assert.Error(err)
	assert.Contains(err.Error(), "condition is required")

	// Should not proceed to Initialize if configuration is incomplete
}

func (suite *TestSuiteCustomSecurity) TestNilReceiver_EnsureConfigurationCompleteness() {
	assert := assert.New(suite.T())

	var spec *CustomSecuritySpec = nil

	// This should panic - testing defensive programming
	assert.Panics(func() {
		spec.EnsureConfigurationCompleteness()
	})
}

func (suite *TestSuiteCustomSecurity) TestNilReceiver_Initialize() {
	assert := assert.New(suite.T())

	var spec *CustomSecuritySpec = nil

	// This should panic - testing defensive programming
	assert.Panics(func() {
		spec.Initialize()
	})
}

func (suite *TestSuiteCustomSecurity) TestNilReceiver_IsSecure() {
	assert := assert.New(suite.T())

	var spec *CustomSecuritySpec = nil

	// This should panic - testing defensive programming
	assert.Panics(func() {
		spec.IsSecure(suite.ctx, suite.requestCtx)
	})
}

func TestRunCustomSecuritySuite(t *testing.T) {
	suite.Run(t, new(TestSuiteCustomSecurity))
}

// Benchmarks

func BenchmarkEnsureConfigurationCompleteness(b *testing.B) {
	condition, _ := valuable.Serialize("true")
	spec := &CustomSecuritySpec{
		Condition: condition,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.EnsureConfigurationCompleteness() // nolint:errcheck
	}
}

func BenchmarkInitialize(b *testing.B) {
	condition, _ := valuable.Serialize("true")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec := &CustomSecuritySpec{
			Condition: condition,
		}
		spec.Initialize() // nolint:errcheck
	}
}

func BenchmarkIsSecure_SimpleCondition(b *testing.B) {
	condition, _ := valuable.Serialize("true")
	spec := &CustomSecuritySpec{
		Condition: condition,
	}
	spec.Initialize() // nolint:errcheck
	
	ctx := context.Background()
	requestCtx := &fasthttpz.RequestCtx{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.IsSecure(ctx, requestCtx) // nolint:errcheck
	}
}

func BenchmarkIsSecure_ComplexCondition(b *testing.B) {
	condition, _ := valuable.Serialize("{{ if eq 1 1 }}true{{ else }}false{{ end }}")
	spec := &CustomSecuritySpec{
		Condition: condition,
	}
	spec.Initialize() // nolint:errcheck
	
	ctx := context.Background()
	requestCtx := &fasthttpz.RequestCtx{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.IsSecure(ctx, requestCtx) // nolint:errcheck
	}
}