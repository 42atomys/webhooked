package valuable

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	os.Setenv(suite.testEnvName, suite.testValue)
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
		input   interface{}
		output  []string
		wantErr bool
	}{
		{"string value", suite.testValue, []string{suite.testValue}, false},
		{"int value", 1, []string{"1"}, false},
		{"float value", 1.42, []string{"1.42"}, false},
		{"boolean value", true, []string{"true"}, false},
		{"map[interface{}]interface{} value", map[interface{}]interface{}{"value": "test"}, []string{"test"}, false},
		{"map[interface{}]interface{} with error", map[interface{}]interface{}{"value": func() {}}, []string{}, true},
		{"nil value", nil, []string{}, false},
		{"simple value map interface", map[string]interface{}{
			"value": suite.testValue,
		}, []string{suite.testValue}, false},
		{"complexe value from envRef map interface", map[string]interface{}{
			"valueFrom": map[string]interface{}{
				"envRef": suite.testEnvName,
			},
		}, []string{suite.testValue}, false},
		{"invalid payload", map[string]interface{}{
			"valueFrom": map[string]interface{}{
				"envRef": func() {},
			},
		}, []string{suite.testValue}, true},
	}

	for _, test := range tests {
		v, err := SerializeValuable(test.input)
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
		{"an environment ref with invalid name", &Valuable{ValueFrom: &ValueFromSource{EnvRef: &suite.testInvalidEnvName}}, []string{""}},
		{"an environment ref with valid name", &Valuable{ValueFrom: &ValueFromSource{EnvRef: &suite.testEnvName}}, []string{suite.testValue}},
		{"a static ref", &Valuable{ValueFrom: &ValueFromSource{StaticRef: &suite.testValue}}, []string{suite.testValue}},
	}

	for _, test := range tests {
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
