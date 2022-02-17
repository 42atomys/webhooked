package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validPort(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		input    int
		expected bool
	}{
		{8080, true},
		{1, true},
		{0, false},
		{-8080, false},
		{65535, false},
		{65536, false},
	}

	for _, test := range tests {
		assert.Equal(validPort(test.input), test.expected, "input: %d", test.input)
	}

}
