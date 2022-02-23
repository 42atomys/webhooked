package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFunctionByName(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		input      string
		expectedOk bool
	}{
		{"getHeader", true},
		{"compareWithStaticValue", true},
		{"invalidFunctionName", false},
	}

	for _, test := range tests {
		fn, ok := GetFunctionByName(test.input)
		assert.Equal(test.expectedOk, ok, "input: %s", test.input)
		if test.expectedOk {
			assert.NotNil(fn, "input: %s", test.input)
		} else {
			assert.Nil(fn, "input: %s", test.input)
		}
	}
}
