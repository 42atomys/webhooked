package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"42stellar.org/webhooks/internal/valuable"
	"42stellar.org/webhooks/pkg/factory"
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
				Security: []map[string]Security{
					{
						"header": Security{"secretHeader", []*factory.InputConfig{
							{
								Name:     "headerName",
								Valuable: valuable.Valuable{Values: []string{"X-Token"}},
							},
						}, make(map[string]interface{})},
						"compare": Security{"", []*factory.InputConfig{
							{
								Name:     "first",
								Valuable: valuable.Valuable{Values: []string{"{{ .Outputs.secretHeader.value }}"}},
							},
							{
								Name:     "second",
								Valuable: valuable.Valuable{Values: []string{"test"}},
							},
						}, map[string]interface{}{"inverse": false}},
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
				Security: []map[string]Security{},
			},
			false,
			0,
		},
		{
			"invalid factory name in configuration",
			&WebhookSpec{
				Name: "test",
				Security: []map[string]Security{
					{
						"invalid": Security{},
					},
				},
			},
			true,
			0,
		},
	}

	for _, test := range tests {
		err := loadSecurityFactory(test.input)
		if test.wantErr {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}
		assert.Equal(test.input.SecurityPipeline.FactoryCount(), test.wantLen)
	}
}

func TestLoadStorage(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name        string
		storageName string
		input       *WebhookSpec
		wantErr     bool
		wantStorage bool
	}{
		{"no spec", "", &WebhookSpec{Name: "test"}, false, false},
		{
			"full valid storage",
			"connection invalid must return an error",
			&WebhookSpec{
				Name: "test",
				Storage: []*StorageSpec{
					{
						Type: "redis",
						Specs: map[string]interface{}{
							"host": "localhost",
							"port": 0,
						},
					},
				},
			},
			true,
			false,
		},
		{
			"empty storage configuration",
			"",
			&WebhookSpec{
				Name:    "test",
				Storage: []*StorageSpec{},
			},
			false,
			false,
		},
		{
			"invalid storage name in configuration",
			"",
			&WebhookSpec{
				Name: "test",
				Storage: []*StorageSpec{
					{},
				},
			},
			true,
			false,
		},
	}

	for _, test := range tests {
		err := loadStorage(test.input)
		if test.wantErr {
			assert.Error(err)
		} else {
			assert.NoError(err)
		}

		if test.wantStorage && assert.Len(test.input.Storage, 1, "no storage is loaded for test %s", test.name) {
			s := test.input.Storage[0]
			assert.NotNil(s)
		}
	}
}
