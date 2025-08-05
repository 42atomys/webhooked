//go:build unit

package noop

import (
	"context"
	"testing"

	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

type TestSuiteNoopSecurity struct {
	suite.Suite

	spec       *NoopSecuritySpec
	ctx        context.Context
	requestCtx *fasthttpz.RequestCtx
}

func (suite *TestSuiteNoopSecurity) BeforeTest(suiteName, testName string) {
	suite.spec = &NoopSecuritySpec{}
	suite.ctx = context.Background()

	// Create a fasthttp request context
	fastCtx := &fasthttp.RequestCtx{}
	suite.requestCtx = &fasthttpz.RequestCtx{RequestCtx: fastCtx}
}

func (suite *TestSuiteNoopSecurity) TestEnsureConfigurationCompleteness() {
	assert := assert.New(suite.T())

	err := suite.spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
}

func (suite *TestSuiteNoopSecurity) TestInitialize() {
	assert := assert.New(suite.T())

	err := suite.spec.Initialize()

	assert.NoError(err)
}

func (suite *TestSuiteNoopSecurity) TestIsSecure() {
	assert := assert.New(suite.T())

	result, err := suite.spec.IsSecure(suite.ctx, suite.requestCtx)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteNoopSecurity) TestIsSecure_WithNilContext() {
	assert := assert.New(suite.T())

	result, err := suite.spec.IsSecure(nil, suite.requestCtx)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteNoopSecurity) TestIsSecure_WithNilRequestCtx() {
	assert := assert.New(suite.T())

	result, err := suite.spec.IsSecure(suite.ctx, nil)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteNoopSecurity) TestIsSecure_WithBothNil() {
	assert := assert.New(suite.T())

	result, err := suite.spec.IsSecure(nil, nil)

	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteNoopSecurity) TestIsSecure_MultipleRequests() {
	assert := assert.New(suite.T())

	// Test multiple calls to ensure consistency
	for i := 0; i < 100; i++ {
		result, err := suite.spec.IsSecure(suite.ctx, suite.requestCtx)

		assert.NoError(err, "Call %d should not error", i)
		assert.True(result, "Call %d should return true", i)
	}
}

func (suite *TestSuiteNoopSecurity) TestIsSecure_DifferentRequestContexts() {
	assert := assert.New(suite.T())

	// Test with different request contexts
	contexts := make([]*fasthttpz.RequestCtx, 5)
	for i := range contexts {
		fastCtx := &fasthttp.RequestCtx{}
		contexts[i] = &fasthttpz.RequestCtx{RequestCtx: fastCtx}

		// Set different request data
		fastCtx.Request.SetRequestURI("https://example.com/webhook")
		fastCtx.Request.Header.SetMethod("POST")
		fastCtx.Request.SetBody([]byte(`{"test": "data"}`))
	}

	for i, requestCtx := range contexts {
		result, err := suite.spec.IsSecure(suite.ctx, requestCtx)

		assert.NoError(err, "Request %d should not error", i)
		assert.True(result, "Request %d should return true", i)
	}
}

func (suite *TestSuiteNoopSecurity) TestFullWorkflow() {
	assert := assert.New(suite.T())

	// Test complete workflow from configuration to security check
	spec := &NoopSecuritySpec{}

	// Step 1: Ensure configuration completeness
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)

	// Step 2: Initialize
	err = spec.Initialize()
	assert.NoError(err)

	// Step 3: Check security
	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)
	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteNoopSecurity) TestNilReceiver_EnsureConfigurationCompleteness() {
	assert := assert.New(suite.T())

	var spec *NoopSecuritySpec = nil

	// Noop methods work with nil receivers since they don't dereference
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)
}

func (suite *TestSuiteNoopSecurity) TestNilReceiver_Initialize() {
	assert := assert.New(suite.T())

	var spec *NoopSecuritySpec = nil

	// Noop methods work with nil receivers since they don't dereference
	err := spec.Initialize()
	assert.NoError(err)
}

func (suite *TestSuiteNoopSecurity) TestNilReceiver_IsSecure() {
	assert := assert.New(suite.T())

	var spec *NoopSecuritySpec = nil

	// Noop methods work with nil receivers since they don't dereference
	result, err := spec.IsSecure(suite.ctx, suite.requestCtx)
	assert.NoError(err)
	assert.True(result)
}

func (suite *TestSuiteNoopSecurity) TestStructInitialization() {
	assert := assert.New(suite.T())

	// Test different ways of creating the struct
	spec1 := &NoopSecuritySpec{}
	spec2 := new(NoopSecuritySpec)
	var spec3 NoopSecuritySpec

	specs := []*NoopSecuritySpec{spec1, spec2, &spec3}

	for i, spec := range specs {
		err := spec.EnsureConfigurationCompleteness()
		assert.NoError(err, "Spec %d should not error on EnsureConfigurationCompleteness", i)

		err = spec.Initialize()
		assert.NoError(err, "Spec %d should not error on Initialize", i)

		result, err := spec.IsSecure(suite.ctx, suite.requestCtx)
		assert.NoError(err, "Spec %d should not error on IsSecure", i)
		assert.True(result, "Spec %d should return true on IsSecure", i)
	}
}

func (suite *TestSuiteNoopSecurity) TestConcurrentAccess() {
	assert := assert.New(suite.T())

	// Test concurrent access to the same spec instance
	spec := &NoopSecuritySpec{}

	// Initialize once
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)

	err = spec.Initialize()
	assert.NoError(err)

	// Run concurrent security checks
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				result, err := spec.IsSecure(suite.ctx, suite.requestCtx)
				assert.NoError(err, "Goroutine %d iteration %d should not error", id, j)
				assert.True(result, "Goroutine %d iteration %d should return true", id, j)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRunNoopSecuritySuite(t *testing.T) {
	suite.Run(t, new(TestSuiteNoopSecurity))
}

// Benchmarks

func BenchmarkEnsureConfigurationCompleteness(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	spec := &NoopSecuritySpec{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.EnsureConfigurationCompleteness() // nolint:errcheck
	}
}

func BenchmarkInitialize(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	spec := &NoopSecuritySpec{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.Initialize() // nolint:errcheck
	}
}

func BenchmarkIsSecure(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	spec := &NoopSecuritySpec{}
	ctx := context.Background()

	fastCtx := &fasthttp.RequestCtx{}
	requestCtx := &fasthttpz.RequestCtx{RequestCtx: fastCtx}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.IsSecure(ctx, requestCtx) // nolint:errcheck
	}
}

func BenchmarkFullWorkflow(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	ctx := context.Background()
	fastCtx := &fasthttp.RequestCtx{}
	requestCtx := &fasthttpz.RequestCtx{RequestCtx: fastCtx}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec := &NoopSecuritySpec{}
		spec.EnsureConfigurationCompleteness() // nolint:errcheck
		spec.Initialize()                      // nolint:errcheck
		spec.IsSecure(ctx, requestCtx)         // nolint:errcheck
	}
}

func BenchmarkConcurrentIsSecure(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	spec := &NoopSecuritySpec{}
	spec.EnsureConfigurationCompleteness() // nolint:errcheck
	spec.Initialize()                      // nolint:errcheck

	ctx := context.Background()
	fastCtx := &fasthttp.RequestCtx{}
	requestCtx := &fasthttpz.RequestCtx{RequestCtx: fastCtx}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spec.IsSecure(ctx, requestCtx) // nolint:errcheck
		}
	})
}
