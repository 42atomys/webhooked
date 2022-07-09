package formatting

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"atomys.codes/webhooked/internal/config"
)

func TestNewTemplateData(t *testing.T) {
	assert := assert.New(t)

	tmpl := NewTemplateData("")
	assert.NotNil(tmpl)
	assert.Equal("", tmpl.tmplString)
	assert.Equal(0, len(tmpl.data))

	tmpl = NewTemplateData("{{ .Payload }}")
	assert.NotNil(tmpl)
	assert.Equal("{{ .Payload }}", tmpl.tmplString)
	assert.Equal(0, len(tmpl.data))
}

func Test_WithData(t *testing.T) {
	assert := assert.New(t)

	tmpl := NewTemplateData("").WithData("test", true)
	assert.NotNil(tmpl)
	assert.Equal("", tmpl.tmplString)
	assert.Equal(1, len(tmpl.data))
	assert.Equal(true, tmpl.data["test"])
}

func Test_WithRequest(t *testing.T) {
	assert := assert.New(t)

	tmpl := NewTemplateData("").WithRequest(httptest.NewRequest("GET", "/", nil))
	assert.NotNil(tmpl)
	assert.Equal("", tmpl.tmplString)
	assert.Equal(1, len(tmpl.data))
	assert.Nil(tmpl.data["request"])
	assert.NotNil(tmpl.data["Request"])
	assert.Equal("GET", tmpl.data["Request"].(*http.Request).Method)
}

func Test_WithPayload(t *testing.T) {
	assert := assert.New(t)

	data, err := json.Marshal(map[string]interface{}{"test": "test"})
	assert.Nil(err)

	tmpl := NewTemplateData("").WithPayload(data)
	assert.NotNil(tmpl)
	assert.Equal("", tmpl.tmplString)
	assert.Equal(1, len(tmpl.data))
	assert.JSONEq(`{"test":"test"}`, tmpl.data["Payload"].(string))
}

func Test_Render(t *testing.T) {
	assert := assert.New(t)

	// Test with basic template
	tmpl := NewTemplateData("{{ .Payload }}").WithPayload([]byte(`{"test": "test"}`))
	assert.NotNil(tmpl)
	assert.Equal("{{ .Payload }}", tmpl.tmplString)
	assert.Equal(1, len(tmpl.data))
	assert.JSONEq(`{"test":"test"}`, tmpl.data["Payload"].(string))

	str, err := tmpl.Render()
	assert.Nil(err)
	assert.JSONEq("{\"test\":\"test\"}", str)

	// Test with template with multiple data sources
	// and complex template
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Test", "test")

	tmpl = NewTemplateData(`
	{
		"config": {{ toJson .Config }},
		"spec": {{ toJson .Spec }},
		"storage": {{ toJson .Storage }},
		"metadata": {
			"testID": "{{ .Request.Header | getHeader "X-Test" }}",
			"deliveryID": "{{ .Request.Header | getHeader "X-Delivery" | default "unknown" }}"
		},
		"payload": {{ .Payload }}
	}
	`).
		WithPayload([]byte(`{"test": "test"}`)).
		WithRequest(req).
		WithData("Spec", &config.WebhookSpec{Name: "test", EntrypointURL: "/webhooks/test", Formatting: &config.FormattingSpec{}}).
		WithData("Storage", &config.StorageSpec{Type: "testing", Specs: map[string]interface{}{}}).
		WithData("Config", config.Current())
	assert.NotNil(tmpl)

	str, err = tmpl.Render()
	assert.Nil(err)
	assert.JSONEq(`{
		"config": {
			"apiVersion":"",
			"observability":{
				"metricsEnabled":false
			},
			"specs": null
		},
		"spec": {
			"name":"test",
			"entrypointUrl": "/webhooks/test"
		},
		"storage": {
			"type": "testing"
		},
		"metadata": {
			"testID": "test",
			"deliveryID": "unknown"
		},
		"payload": {
			"test": "test"
		}
	}`, str)

	// Test with template with template error
	tmpl = NewTemplateData("{{ .Payload }")
	assert.NotNil(tmpl)
	assert.Equal("{{ .Payload }", tmpl.tmplString)

	str, err = tmpl.Render()
	assert.Error(err)
	assert.Contains(err.Error(), "error in your template: ")
	assert.Equal("", str)

	// Test with template with data error
	tmpl = NewTemplateData("{{ .Request.Method }}").WithRequest(nil)
	assert.NotNil(tmpl)
	assert.Equal("{{ .Request.Method }}", tmpl.tmplString)

	str, err = tmpl.Render()
	assert.Error(err)
	assert.Contains(err.Error(), "error while filling your template: ")
	assert.Equal("", str)
}
