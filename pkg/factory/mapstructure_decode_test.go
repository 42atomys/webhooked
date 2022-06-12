package factory

import (
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteInputConfigDecode struct {
	suite.Suite

	testValue, testName string
	testInputConfig     map[interface{}]interface{}

	decodeFunc func(input, output interface{}) (err error)
}

func (suite *TestSuiteInputConfigDecode) BeforeTest(suiteName, testName string) {
	suite.testName = "testName"
	suite.testValue = "testValue"
	suite.testInputConfig = map[interface{}]interface{}{
		"name":  suite.testName,
		"value": suite.testValue,
	}

	suite.decodeFunc = func(input, output interface{}) (err error) {
		var decoder *mapstructure.Decoder

		decoder, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:     output,
			DecodeHook: DecodeHook,
		})
		if err != nil {
			return err
		}

		return decoder.Decode(input)
	}
}

func (suite *TestSuiteInputConfigDecode) TestDecodeInvalidOutput() {
	assert := assert.New(suite.T())

	err := suite.decodeFunc(map[interface{}]interface{}{"value": suite.testValue}, nil)
	assert.Error(err)
}

func (suite *TestSuiteInputConfigDecode) TestDecodeInvalidInput() {
	assert := assert.New(suite.T())

	output := struct{}{}
	err := suite.decodeFunc(map[interface{}]interface{}{"value": true}, &output)
	assert.NoError(err)
}

func (suite *TestSuiteInputConfigDecode) TestDecodeString() {
	assert := assert.New(suite.T())

	output := InputConfig{}
	err := suite.decodeFunc(suite.testInputConfig, &output)
	assert.NoError(err)
	assert.Equal(suite.testName, output.Name)
	assert.Equal(suite.testValue, output.First())
}

func TestRunSuiteInputConfigDecode(t *testing.T) {
	suite.Run(t, new(TestSuiteInputConfigDecode))
}
