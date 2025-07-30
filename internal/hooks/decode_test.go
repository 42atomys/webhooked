//go:build unit

package hooks

import (
	"testing"

	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteDecodeField struct {
	suite.Suite

	validMapData   map[string]any
	invalidData    map[string]any
	testKey        string
	nonExistentKey string
}

func (suite *TestSuiteDecodeField) BeforeTest(suiteName, testName string) {
	suite.testKey = "testField"
	suite.nonExistentKey = "nonExistent"
	suite.validMapData = map[string]any{
		suite.testKey: map[string]any{
			"value":  "testValue",
			"values": []string{"val1", "val2"},
		},
	}
	suite.invalidData = map[string]any{
		suite.testKey: "not a map",
	}
}

func (suite *TestSuiteDecodeField) TestDecodeFieldKeyNotExists() {
	assert := assert.New(suite.T())

	type result struct {
		Value string `json:"value"`
	}

	var output result
	err := DecodeField(suite.validMapData, suite.nonExistentKey, &output)
	assert.NoError(err)
	assert.Empty(output.Value)
}

func (suite *TestSuiteDecodeField) TestDecodeFieldInvalidMapType() {
	assert := assert.New(suite.T())

	type result struct {
		Value string `json:"value"`
	}

	var output result
	err := DecodeField(suite.invalidData, suite.testKey, &output)
	assert.Error(err)
	assert.Contains(err.Error(), "must be a map")
}

func (suite *TestSuiteDecodeField) TestDecodeFieldValidDecode() {
	assert := assert.New(suite.T())

	type result struct {
		Value  string   `json:"value"`
		Values []string `json:"values"`
	}

	var output result
	err := DecodeField(suite.validMapData, suite.testKey, &output)
	assert.NoError(err)
	assert.Equal("testValue", output.Value)
	assert.Equal([]string{"val1", "val2"}, output.Values)
}

func (suite *TestSuiteDecodeField) TestDecodeFieldWithValuable() {
	assert := assert.New(suite.T())

	type result struct {
		Value valuable.Valuable `json:"value"`
	}

	var output result
	err := DecodeField(suite.validMapData, suite.testKey, &output)
	assert.NoError(err)
	assert.Equal("testValue", output.Value.First())
}

func (suite *TestSuiteDecodeField) TestDecodeFieldNilResult() {
	assert := assert.New(suite.T())

	err := DecodeField(suite.validMapData, suite.testKey, nil)
	assert.Error(err)
}

func TestRunSuiteDecodeField(t *testing.T) {
	suite.Run(t, new(TestSuiteDecodeField))
}
