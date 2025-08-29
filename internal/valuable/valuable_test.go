//go:build unit

package valuable

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuiteValuable struct {
	suite.Suite

	testValue          string
	testValues         []string
	testEnvName        string
	testInvalidEnvName string
}

func (suite *TestSuiteValuable) BeforeTest(suiteName, testName string) {
	suite.testValue = "test"
	suite.testValues = []string{"test1", "test2"}
	suite.testEnvName = "TEST_WEBHOOKED_CONFIG_ENVREF"
	suite.testInvalidEnvName = "TEST_WEBHOOKED_CONFIG_ENVREF_INVALID"
	require.NoError(suite.T(), os.Setenv(suite.testEnvName, suite.testValue))
}

func (suite *TestSuiteValuable) TestValidate() {
	assert := assert.New(suite.T())

	tests := []struct {
		name    string
		input   *Valuable
		wantErr bool
	}{
		{"a basic value", &Valuable{Value: &suite.testValue}, false},
		{"a basic list of values", &Valuable{Values: suite.testValues}, false},
		{"a basic value with a basic list", &Valuable{Value: &suite.testValue, Values: suite.testValues}, false},
		{"an empty valueFrom", &Valuable{ValueFrom: &ValueFromSource{}}, false},
		{"an environment ref with invalid name", &Valuable{ValueFrom: &ValueFromSource{EnvRef: &suite.testInvalidEnvName}}, true},
		{"an environment ref with valid name", &Valuable{ValueFrom: &ValueFromSource{EnvRef: &suite.testEnvName}}, false},
	}

	for _, test := range tests {
		err := test.input.Validate()
		if test.wantErr && assert.Error(err, "this test must be crash %s", err) {
		} else {
			assert.NoError(err, "cannot validate test %s", test.name)
		}
	}
}

func (suite *TestSuiteValuable) TestSerializeValuable() {
	assert := assert.New(suite.T())

	tests := []struct {
		name    string
		input   any
		output  []string
		wantErr bool
	}{
		{"string value", suite.testValue, []string{suite.testValue}, false},
		{"int value", 1, []string{"1"}, false},
		{"float value", 1.42, []string{"1.42"}, false},
		{"boolean value", true, []string{"true"}, false},
		{"map[any]any value", map[any]any{"value": "test"}, []string{"test"}, false},
		{"map[any]any with error", map[any]any{"value": func() {}}, []string{}, true},
		{"nil value", nil, []string{}, false},
		{"simple value map interface", map[string]any{
			"value": suite.testValue,
		}, []string{suite.testValue}, false},
		{"complexe value from envRef map interface", map[string]any{
			"valueFrom": map[string]any{
				"envRef": suite.testEnvName,
			},
		}, []string{suite.testValue}, false},
		{"invalid payload", map[string]any{
			"valueFrom": map[string]any{
				"envRef": func() {},
			},
		}, []string{suite.testValue}, true},
	}

	for _, test := range tests {
		v, err := Serialize(test.input)
		if test.wantErr && assert.Error(err, "this test must be crash %s", err) {
		} else if assert.NoError(err, "cannot serialize test %s", test.name) {
			assert.ElementsMatch(v.Get(), test.output, test.name)
		}
	}
}

func (suite *TestSuiteValuable) TestValuableGet() {
	assert := assert.New(suite.T())

	tests := []struct {
		name   string
		input  *Valuable
		output []string
	}{
		{"a basic value", &Valuable{Value: &suite.testValue}, []string{suite.testValue}},
		{"a basic list of values", &Valuable{Values: suite.testValues}, suite.testValues},
		{"a basic value with a basic list", &Valuable{Value: &suite.testValue, Values: suite.testValues}, append(suite.testValues, suite.testValue)},
		{"an empty valueFrom", &Valuable{ValueFrom: &ValueFromSource{}}, []string{}},
		{"an environment ref with invalid name", &Valuable{ValueFrom: &ValueFromSource{EnvRef: &suite.testInvalidEnvName}}, []string{}},
		{"an environment ref with valid name", &Valuable{ValueFrom: &ValueFromSource{EnvRef: &suite.testEnvName}}, []string{suite.testValue}},
		{"a static ref", &Valuable{ValueFrom: &ValueFromSource{StaticRef: &suite.testValue}}, []string{suite.testValue}},
	}

	for _, test := range tests {
		assert.NoError(test.input.retrieveData())
		assert.ElementsMatch(test.input.Get(), test.output, test.name)
	}
}

func (suite *TestSuiteValuable) TestValuableFirstandString() {
	assert := assert.New(suite.T())

	tests := []struct {
		name   string
		input  *Valuable
		output string
	}{
		{"a basic value", &Valuable{Value: &suite.testValue}, suite.testValue},
		{"a basic list of values", &Valuable{Values: suite.testValues}, suite.testValues[0]},
		{"a basic value with a basic list", &Valuable{Value: &suite.testValue, Values: suite.testValues}, suite.testValues[0]},
		{"an empty valueFrom", &Valuable{ValueFrom: &ValueFromSource{}}, ""},
		{"an environment ref with invalid name", &Valuable{ValueFrom: &ValueFromSource{EnvRef: &suite.testInvalidEnvName}}, ""},
		{"an environment ref with valid name", &Valuable{ValueFrom: &ValueFromSource{EnvRef: &suite.testEnvName}}, suite.testValue},
		{"a static ref", &Valuable{ValueFrom: &ValueFromSource{StaticRef: &suite.testValue}}, suite.testValue},
	}

	for _, test := range tests {
		assert.NoError(test.input.retrieveData())
		assert.Equal(test.input.First(), test.output, test.name)
		assert.Equal(test.input.String(), test.output, test.name)
	}
}

func (suite *TestSuiteValuable) TestValuableContains() {
	assert := assert.New(suite.T())

	tests := []struct {
		name       string
		input      []string
		testString string
		output     bool
	}{
		{"with nil list", nil, suite.testValue, false},
		{"with nil value", nil, suite.testValue, false},
		{"with empty list", []string{}, suite.testValue, false},
		{"with not included value", []string{"invalid"}, suite.testValue, false},
		{"with included value", []string{suite.testValue}, suite.testValue, true},
	}

	for _, test := range tests {
		v := Valuable{Values: test.input}
		assert.NoError(v.retrieveData(), test.name)
		assert.Equal(test.output, v.Contains(test.testString), test.name)
	}
}

func (suite *TestSuiteValuable) TestValuablecontains() {
	assert := assert.New(suite.T())

	tests := []struct {
		name       string
		input      []string
		testString string
		output     bool
	}{
		{"with nil list", nil, suite.testValue, false},
		{"with nil value", nil, suite.testValue, false},
		{"with empty list", []string{}, suite.testValue, false},
		{"with not included value", []string{"invalid"}, suite.testValue, false},
		{"with included value", []string{suite.testValue}, suite.testValue, true},
	}

	for _, test := range tests {
		v := Valuable{Values: test.input}
		assert.NoError(v.retrieveData(), test.name)
		assert.Equal(test.output, contains(v.Get(), test.testString), test.name)
	}
}

func (suite *TestSuiteValuable) TestValuablecommaListIfAbsent() {
	assert := assert.New(suite.T())

	tests := []struct {
		name   string
		input  string
		output []string
	}{
		{"with uniq list", "foo,bar", []string{"foo", "bar"}},
		{"with no uniq list", "foo,foo,bar", []string{"foo", "bar"}},
	}

	for _, test := range tests {
		assert.Equal(test.output, appendCommaListIfAbsent([]string{}, test.input), test.name)
	}
}

func TestRunValuableSuite(t *testing.T) {
	suite.Run(t, new(TestSuiteValuable))
}

// Benchmarks

func BenchmarkValuable_Get(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	testValue := "test"
	v := &Valuable{Value: &testValue}
	err := v.retrieveData()
	require.NoError(b, err, "Failed to retrieve data for benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Get()
	}
}

func BenchmarkValuable_Get_WithValues(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	v := &Valuable{Values: []string{"test1", "test2", "test3", "test4", "test5"}}
	err := v.retrieveData()
	require.NoError(b, err, "Failed to retrieve data for benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Get()
	}
}

func BenchmarkValuable_Get_WithEnvRef(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	envName := "BENCH_TEST_ENV"
	os.Setenv(envName, "benchvalue") // nolint:errcheck
	defer os.Unsetenv(envName)       // nolint:errcheck

	v := &Valuable{ValueFrom: &ValueFromSource{EnvRef: &envName}}
	err := v.retrieveData()
	require.NoError(b, err, "Failed to retrieve data for benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Get()
	}
}

func BenchmarkValuable_Contains(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	v := &Valuable{Values: []string{"test1", "test2", "test3", "test4", "test5"}}
	err := v.retrieveData()
	require.NoError(b, err, "Failed to retrieve data for benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Contains("test3")
	}
}

func BenchmarkValuable_Contains_NotFound(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	v := &Valuable{Values: []string{"test1", "test2", "test3", "test4", "test5"}}
	err := v.retrieveData()
	require.NoError(b, err, "Failed to retrieve data for benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Contains("notfound")
	}
}

func BenchmarkSerialize_String(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	testValue := "test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Serialize(testValue) // nolint:errcheck
	}
}

func BenchmarkSerialize_Map(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	testMap := map[string]any{
		"value": "test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Serialize(testMap) // nolint:errcheck
	}
}

func BenchmarkSerialize_ComplexMap(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	testMap := map[string]any{
		"valueFrom": map[string]any{
			"envRef": "TEST_ENV",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Serialize(testMap) // nolint:errcheck
	}
}

func BenchmarkValuable_Validate(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	testValue := "test"
	v := &Valuable{Value: &testValue}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate() // nolint:errcheck
	}
}

func BenchmarkAppendCommaListIfAbsent(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		appendCommaListIfAbsent([]string{}, "foo,bar,baz,qux")
	}
}

func BenchmarkAppendCommaListIfAbsent_WithDuplicates(b *testing.B) {
	log.Logger = log.Output(zerolog.Nop())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		appendCommaListIfAbsent([]string{}, "foo,foo,bar,bar,baz,baz")
	}
}
