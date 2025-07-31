//go:build unit

package noop

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteNoopStorage struct {
	suite.Suite

	spec *NoopStorageSpec
	ctx  context.Context
}

func (suite *TestSuiteNoopStorage) BeforeTest(suiteName, testName string) {
	suite.spec = &NoopStorageSpec{}
	suite.ctx = context.Background()
}

func (suite *TestSuiteNoopStorage) TestEnsureConfigurationCompleteness() {
	assert := assert.New(suite.T())

	err := suite.spec.EnsureConfigurationCompleteness()

	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestInitialize() {
	assert := assert.New(suite.T())

	err := suite.spec.Initialize()

	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestStore() {
	assert := assert.New(suite.T())

	testData := []byte(`{"test": "data", "value": 123}`)
	err := suite.spec.Store(suite.ctx, testData)

	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestStore_WithNilContext() {
	assert := assert.New(suite.T())

	testData := []byte(`{"test": "data"}`)
	err := suite.spec.Store(nil, testData)

	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestStore_WithNilData() {
	assert := assert.New(suite.T())

	err := suite.spec.Store(suite.ctx, nil)

	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestStore_WithEmptyData() {
	assert := assert.New(suite.T())

	err := suite.spec.Store(suite.ctx, []byte{})

	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestStore_WithBothNil() {
	assert := assert.New(suite.T())

	err := suite.spec.Store(nil, nil)

	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestStore_MultipleOperations() {
	assert := assert.New(suite.T())

	testData := [][]byte{
		[]byte(`{"message": "first"}`),
		[]byte(`{"message": "second"}`),
		[]byte(`{"message": "third"}`),
		[]byte(`<xml><message>fourth</message></xml>`),
		[]byte(`plain text message`),
	}

	for i, data := range testData {
		err := suite.spec.Store(suite.ctx, data)
		assert.NoError(err, "Store operation %d should not error", i)
	}
}

func (suite *TestSuiteNoopStorage) TestStore_LargeData() {
	assert := assert.New(suite.T())

	// Create large data (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte('A' + (i % 26))
	}

	err := suite.spec.Store(suite.ctx, largeData)

	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestStore_DifferentDataTypes() {
	assert := assert.New(suite.T())

	testCases := []struct {
		name string
		data []byte
	}{
		{"JSON", []byte(`{"key": "value", "number": 42}`)},
		{"XML", []byte(`<root><element>value</element></root>`)},
		{"Plain Text", []byte(`This is plain text`)},
		{"Binary", []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}},
		{"Unicode", []byte(`{"message": "Hello 世界 🌍"}`)},
		{"Empty", []byte{}},
		{"Single Byte", []byte{0x42}},
	}

	for _, tc := range testCases {
		err := suite.spec.Store(suite.ctx, tc.data)
		assert.NoError(err, "Store operation for %s should not error", tc.name)
	}
}

func (suite *TestSuiteNoopStorage) TestFullWorkflow() {
	assert := assert.New(suite.T())

	// Test complete workflow from configuration to storage
	spec := &NoopStorageSpec{}

	// Step 1: Ensure configuration completeness
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)

	// Step 2: Initialize
	err = spec.Initialize()
	assert.NoError(err)

	// Step 3: Store data
	testData := []byte(`{"workflow": "test"}`)
	err = spec.Store(suite.ctx, testData)
	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestNilReceiver_EnsureConfigurationCompleteness() {
	assert := assert.New(suite.T())

	var spec *NoopStorageSpec = nil

	// Noop methods work with nil receivers since they don't dereference
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestNilReceiver_Initialize() {
	assert := assert.New(suite.T())

	var spec *NoopStorageSpec = nil

	// Noop methods work with nil receivers since they don't dereference
	err := spec.Initialize()
	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestNilReceiver_Store() {
	assert := assert.New(suite.T())

	var spec *NoopStorageSpec = nil

	// Noop methods work with nil receivers since they don't dereference
	err := spec.Store(suite.ctx, []byte("test"))
	assert.NoError(err)
}

func (suite *TestSuiteNoopStorage) TestStructInitialization() {
	assert := assert.New(suite.T())

	// Test different ways of creating the struct
	spec1 := &NoopStorageSpec{}
	spec2 := new(NoopStorageSpec)
	var spec3 NoopStorageSpec

	specs := []*NoopStorageSpec{spec1, spec2, &spec3}
	testData := []byte(`{"init": "test"}`)

	for i, spec := range specs {
		err := spec.EnsureConfigurationCompleteness()
		assert.NoError(err, "Spec %d should not error on EnsureConfigurationCompleteness", i)

		err = spec.Initialize()
		assert.NoError(err, "Spec %d should not error on Initialize", i)

		err = spec.Store(suite.ctx, testData)
		assert.NoError(err, "Spec %d should not error on Store", i)
	}
}

func (suite *TestSuiteNoopStorage) TestConcurrentAccess() {
	assert := assert.New(suite.T())

	// Test concurrent access to the same spec instance
	spec := &NoopStorageSpec{}
	
	// Initialize once
	err := spec.EnsureConfigurationCompleteness()
	assert.NoError(err)
	
	err = spec.Initialize()
	assert.NoError(err)

	// Run concurrent store operations
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			for j := 0; j < 10; j++ {
				testData := []byte(`{"goroutine": ` + string(rune('0'+id)) + `, "iteration": ` + string(rune('0'+j)) + `}`)
				err := spec.Store(suite.ctx, testData)
				assert.NoError(err, "Goroutine %d iteration %d should not error", id, j)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func (suite *TestSuiteNoopStorage) TestContextCancellation() {
	assert := assert.New(suite.T())

	// Test with cancelled context
	ctx, cancel := context.WithCancel(suite.ctx)
	cancel() // Cancel immediately

	testData := []byte(`{"cancelled": true}`)
	err := suite.spec.Store(ctx, testData)

	// Noop storage should not care about context cancellation
	assert.NoError(err)
}

func TestRunNoopStorageSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteNoopStorage))
}

// Benchmarks

func BenchmarkEnsureConfigurationCompleteness(b *testing.B) {
	spec := &NoopStorageSpec{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.EnsureConfigurationCompleteness() // nolint:errcheck
	}
}

func BenchmarkInitialize(b *testing.B) {
	spec := &NoopStorageSpec{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.Initialize() // nolint:errcheck
	}
}

func BenchmarkStore(b *testing.B) {
	spec := &NoopStorageSpec{}
	ctx := context.Background()
	testData := []byte(`{"benchmark": "data"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.Store(ctx, testData) // nolint:errcheck
	}
}

func BenchmarkStore_LargeData(b *testing.B) {
	spec := &NoopStorageSpec{}
	ctx := context.Background()
	
	// Create 1MB of data
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte('A' + (i % 26))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec.Store(ctx, largeData) // nolint:errcheck
	}
}

func BenchmarkFullWorkflow(b *testing.B) {
	ctx := context.Background()
	testData := []byte(`{"benchmark": "workflow"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec := &NoopStorageSpec{}
		spec.EnsureConfigurationCompleteness() // nolint:errcheck
		spec.Initialize()                      // nolint:errcheck
		spec.Store(ctx, testData)              // nolint:errcheck
	}
}

func BenchmarkConcurrentStore(b *testing.B) {
	spec := &NoopStorageSpec{}
	spec.EnsureConfigurationCompleteness() // nolint:errcheck
	spec.Initialize()                      // nolint:errcheck
	
	ctx := context.Background()
	testData := []byte(`{"concurrent": "benchmark"}`)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			spec.Store(ctx, testData) // nolint:errcheck
		}
	})
}