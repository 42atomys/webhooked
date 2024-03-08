package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"atomys.codes/webhooked/internal/valuable"
	"atomys.codes/webhooked/pkg/factory"
)

func TestLoad(t *testing.T) {
	os.Setenv("WH_APIVERSION", "v1alpha1_test")
	assert := assert.New(t)
	assert.NoError(Load("../../tests/webhooks.tests.yaml"))

	assert.Equal(true, currentConfig.Observability.MetricsEnabled)
	assert.Equal("v1alpha1_test", currentConfig.APIVersion)
	assert.Len(currentConfig.Specs, 1)

	currentSpec := currentConfig.Specs[0]
	assert.Equal("exampleHook", currentSpec.Name)
	assert.Equal("/webhooks/example", currentSpec.EntrypointURL)

	// Security block
	assert.True(currentSpec.HasSecurity())
	assert.Len(currentSpec.Security, 2)

	// Formating block
	assert.True(currentSpec.HasGlobalFormatting())
	assert.NotEmpty(currentSpec.Formatting.TemplateString)

	// Storage block
	assert.Len(currentSpec.Storage, 1)
	assert.Equal("postgres", currentSpec.Storage[0].Type)
	assert.NotEmpty("postgres", currentSpec.Storage[0].Specs["args"])
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
			assert.Error(err, test.name)
		} else {
			assert.NoError(err, test.name)
		}
		assert.Equal(test.input.SecurityPipeline.FactoryCount(), test.wantLen, test.name)
	}
}

func TestLoadStorage(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name        string
		input       *WebhookSpec
		wantErr     bool
		wantStorage bool
	}{
		{"no spec", &WebhookSpec{Name: "test"}, false, false},
		{
			"full valid storage",
			&WebhookSpec{
				Name: "test",
				Storage: []*StorageSpec{
					{
						Type: "redis",
						Specs: map[string]interface{}{
							"host": "localhost",
							"port": 0,
						},
						Formatting: &FormattingSpec{TemplateString: "null"},
					},
				},
			},
			true,
			false,
		},
		{
			"empty storage configuration",
			&WebhookSpec{
				Name:    "test",
				Storage: []*StorageSpec{},
			},
			false,
			false,
		},
		{
			"invalid storage name in configuration",
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
			assert.Error(err, test.name)
		} else {
			assert.NoError(err, test.name)
		}

		if test.wantStorage && assert.Len(test.input.Storage, 1, "no storage is loaded for test %s", test.name) {
			s := test.input.Storage[0]
			assert.NotNil(s, test.name)
		}
	}
}

func Test_loadTemplate(t *testing.T) {
	tests := []struct {
		name         string
		input        *FormattingSpec
		parentSpec   *FormattingSpec
		wantErr      bool
		wantTemplate string
	}{
		{
			"no template",
			nil,
			nil,
			false,
			defaultPayloadTemplate,
		},
		{
			"template string",
			&FormattingSpec{TemplateString: "{{ .Request.Method }}"},
			nil,
			false,
			"{{ .Request.Method }}",
		},
		{
			"template file",
			&FormattingSpec{TemplatePath: "../../tests/simple_template.tpl"},
			nil,
			false,
			"{{ .Request.Method }}",
		},
		{
			"template file with template string",
			&FormattingSpec{TemplatePath: "../../tests/simple_template.tpl", TemplateString: "{{ .Request.Path }}"},
			nil,
			false,
			"{{ .Request.Path }}",
		},
		{
			"no template with not loaded parent",
			nil,
			&FormattingSpec{TemplateString: "{{ .Request.Method }}"},
			false,
			"{{ .Request.Method }}",
		},
		{
			"no template with loaded parent",
			nil,
			&FormattingSpec{Template: "{{ .Request.Method }}", TemplateString: "{{ .Request.Path }}"},
			false,
			"{{ .Request.Method }}",
		},
		{
			"no template with unloaded parent and error",
			nil,
			&FormattingSpec{TemplatePath: "//invalid//path//"},
			true,
			"",
		},
		{
			"template file not found",
			&FormattingSpec{TemplatePath: "//invalid//path//"},
			nil,
			true,
			"",
		},
	}

	for _, test := range tests {
		tmpl, err := loadTemplate(test.input, test.parentSpec, defaultPayloadTemplate)
		if test.wantErr {
			assert.Error(t, err, test.name)
		} else {
			assert.NoError(t, err, test.name)
		}
		assert.NotNil(t, tmpl, test.name)
		assert.Equal(t, test.wantTemplate, tmpl.Template, test.name)
	}
}
