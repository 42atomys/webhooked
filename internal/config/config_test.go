package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	assert.Equal(t, nil, Validate())
}

func TestCurrent(t *testing.T) {
	assert.Equal(t, config, Current())
}

func TestConfiguration_GetEntry(t *testing.T) {
	var c = &Configuration{Specs: make(map[string]ConfigurationSpec)}
	entry, err := c.GetEntry("missing")
	assert.Equal(t, ErrSpecNotFound, err)
	assert.Equal(t, (*ConfigurationSpec)(nil), entry)

	var testSpec = ConfigurationSpec{
		EntrypointURL: "/test",
	}
	c.Specs["test"] = testSpec

	entry, err = c.GetEntry("test")
	assert.Equal(t, nil, err)
	assert.Equal(t, &testSpec, entry)
}

func TestConfiguration_GetEntryByEndpoint(t *testing.T) {
	var c = &Configuration{Specs: make(map[string]ConfigurationSpec)}
	entry, err := c.GetEntryByEndpoint("/test")
	assert.Equal(t, ErrSpecNotFound, err)
	assert.Equal(t, (*ConfigurationSpec)(nil), entry)

	var testSpec = ConfigurationSpec{
		EntrypointURL: "/test",
	}
	c.Specs["test"] = testSpec

	entry, err = c.GetEntryByEndpoint("/test")
	assert.Equal(t, nil, err)
	assert.Equal(t, &testSpec, entry)
}
