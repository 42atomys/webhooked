package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfiguration(t *testing.T) {
	assert.Equal(t, nil, ValidateConfiguration())
}

func TestGetConfig(t *testing.T) {
	assert.Equal(t, config, GetConfig())
}
