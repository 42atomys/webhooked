package valuable

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteValuableDecode struct {
	suite.Suite

	testValue, testValueCommaSeparated string
	testValues                         []string
}

func (suite *TestSuiteValuableDecode) BeforeTest(suiteName, testName string) {
	suite.testValue = "testValue"
	suite.testValues = []string{"testValue1", "testValue2"}
	suite.testValueCommaSeparated = "testValue3,testValue4"
}

func (suite *TestSuiteValuableDecode) TestDecodeInvalidOutput() {
	assert := assert.New(suite.T())

	err := Decode(map[string]interface{}{"value": suite.testValue}, nil)
	assert.Error(err)
}

func (suite *TestSuiteValuableDecode) TestDecodeString() {
	assert := assert.New(suite.T())

	type strukt struct {
		Value string `mapstructure:"value"`
	}

	output := strukt{}
	err := Decode(map[string]interface{}{"value": suite.testValue}, &output)
	assert.NoError(err)
	assert.Equal(suite.testValue, output.Value)
}

func (suite *TestSuiteValuableDecode) TestDecodeValuableRootString() {
	assert := assert.New(suite.T())

	type strukt struct {
		Value Valuable `mapstructure:"value"`
	}

	output := strukt{}
	err := Decode(map[string]interface{}{"value": suite.testValue}, &output)
	assert.NoError(err)
	assert.Equal(suite.testValue, output.Value.First())
}

func (suite *TestSuiteValuableDecode) TestDecodeValuableRootBool() {
	assert := assert.New(suite.T())

	type strukt struct {
		Value Valuable `mapstructure:"value"`
	}

	output := strukt{}
	err := Decode(map[string]interface{}{"value": true}, &output)
	assert.NoError(err)
	assert.Equal("true", output.Value.First())
}

func (suite *TestSuiteValuableDecode) TestDecodeValuableValue() {
	assert := assert.New(suite.T())

	type strukt struct {
		Value Valuable `mapstructure:"value"`
	}

	output := strukt{}
	err := Decode(map[string]interface{}{"value": map[string]interface{}{"value": suite.testValue}}, &output)
	assert.NoError(err)
	assert.Equal(suite.testValue, output.Value.First())
}

func (suite *TestSuiteValuableDecode) TestDecodeValuableValues() {
	assert := assert.New(suite.T())

	type strukt struct {
		Value Valuable `mapstructure:"value"`
	}

	output := strukt{}
	err := Decode(map[string]interface{}{"value": map[string]interface{}{"values": suite.testValues}}, &output)
	assert.NoError(err)
	assert.Equal(suite.testValues, output.Value.Get())
}

func (suite *TestSuiteValuableDecode) TestDecodeValuableStaticValuesWithComma() {
	assert := assert.New(suite.T())

	type strukt struct {
		Value Valuable `mapstructure:"value"`
	}

	output := strukt{}
	err := Decode(map[string]interface{}{"value": map[string]interface{}{"valueFrom": map[string]interface{}{"staticRef": suite.testValueCommaSeparated}}}, &output)
	assert.NoError(err)
	assert.Equal(strings.Split(suite.testValueCommaSeparated, ","), output.Value.Get())
}

func TestRunSuiteValuableDecode(t *testing.T) {
	suite.Run(t, new(TestSuiteValuableDecode))
}
