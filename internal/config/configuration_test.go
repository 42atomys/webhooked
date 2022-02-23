package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	assert.NoError(t, Load())
}

func TestValidate(t *testing.T) {
	assert.NoError(t, Validate(&Configuration{}))
	assert.NoError(t, Validate(&Configuration{
		Specs: []*WebhookSpec{
			{
				Name:          "test",
				EntrypointURL: "/test",
			},
		},
	}))

	assert.Error(t, Validate(&Configuration{
		Specs: []*WebhookSpec{
			{
				Name:          "test",
				EntrypointURL: "/test",
			},
			{
				Name:          "test2",
				EntrypointURL: "/test",
			},
		},
	}))

	assert.Error(t, Validate(&Configuration{
		Specs: []*WebhookSpec{
			{
				Name:          "test",
				EntrypointURL: "/test",
			},
			{
				Name:          "test",
				EntrypointURL: "/test",
			},
		},
	}))
}

func TestCurrent(t *testing.T) {
	assert.Equal(t, currentConfig, Current())
}

func TestConfiguration_GetSpec(t *testing.T) {
	var c = &Configuration{Specs: make([]*WebhookSpec, 0)}
	spec, err := c.GetSpec("missing")
	assert.Equal(t, ErrSpecNotFound, err)
	assert.Equal(t, (*WebhookSpec)(nil), spec)

	var testSpec = WebhookSpec{
		Name:          "test",
		EntrypointURL: "/test",
	}
	c.Specs = append(c.Specs, &testSpec)

	spec, err = c.GetSpec("test")
	assert.Equal(t, nil, err)
	assert.Equal(t, &testSpec, spec)
}

func TestConfiguration_GeSpecByEndpoint(t *testing.T) {
	var c = &Configuration{Specs: make([]*WebhookSpec, 0)}
	spec, err := c.GetSpecByEndpoint("/test")
	assert.Equal(t, ErrSpecNotFound, err)
	assert.Equal(t, (*WebhookSpec)(nil), spec)

	var testSpec = WebhookSpec{
		EntrypointURL: "/test",
	}
	c.Specs = append(c.Specs, &testSpec)

	spec, err = c.GetSpecByEndpoint("/test")
	assert.Equal(t, nil, err)
	assert.Equal(t, &testSpec, spec)
}

func TestLoadSecurityFactory(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name    string
		input   *WebhookSpec
		wantErr bool
		wantLen int
	}{
		{"no spec", &WebhookSpec{Name: "test"}, false, 0},
		{
			"full valid security",
			&WebhookSpec{
				Name: "test",
				Security: []map[string]map[string]interface{}{
					{
						"getHeader": map[string]interface{}{
							"name": "X-Token",
						},
						"compareWithStaticValue": map[string]interface{}{
							"value": "test",
						},
					},
				},
			},
			false,
			2,
		},
		{
			"empty security configuration",
			&WebhookSpec{
				Name:     "test",
				Security: []map[string]map[string]interface{}{},
			},
			false,
			0,
		},
		{
			"invalid factory name in configuration",
			&WebhookSpec{
				Name: "test",
				Security: []map[string]map[string]interface{}{
					{
						"invalid": map[string]interface{}{},
					},
				},
			},
			true,
			0,
		},
	}

	for _, test := range tests {
		err := LoadSecurityFactory(test.input)
		if test.wantErr {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
		assert.Len(test.input.SecurityFactories, test.wantLen)
	}
}
