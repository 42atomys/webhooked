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

func TestConfiguration_GetSpec(t *testing.T) {
	var c = &Configuration{Specs: make(map[string]WebhookSpec)}
	spec, err := c.GetSpec("missing")
	assert.Equal(t, ErrSpecNotFound, err)
	assert.Equal(t, (*WebhookSpec)(nil), spec)

	var testSpec = WebhookSpec{
		EntrypointURL: "/test",
	}
	c.Specs["test"] = testSpec

	spec, err = c.GetSpec("test")
	assert.Equal(t, nil, err)
	assert.Equal(t, &testSpec, spec)
}

func TestConfiguration_GeSpecByEndpoint(t *testing.T) {
	var c = &Configuration{Specs: make(map[string]WebhookSpec)}
	spec, err := c.GetSpecByEndpoint("/test")
	assert.Equal(t, ErrSpecNotFound, err)
	assert.Equal(t, (*WebhookSpec)(nil), spec)

	var testSpec = WebhookSpec{
		EntrypointURL: "/test",
	}
	c.Specs["test"] = testSpec

	spec, err = c.GetSpecByEndpoint("/test")
	assert.Equal(t, nil, err)
	assert.Equal(t, &testSpec, spec)
}
