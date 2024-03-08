package formatting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWithTemplate(t *testing.T) {
	assert := assert.New(t)

	tmpl := New().WithTemplate("")
	assert.NotNil(tmpl)
	assert.Equal("", tmpl.tmplString)
	assert.Equal(0, len(tmpl.data))

	tmpl = New().WithTemplate("{{ .Payload }}")
	assert.NotNil(tmpl)
	assert.Equal("{{ .Payload }}", tmpl.tmplString)
	assert.Equal(0, len(tmpl.data))

	tmpl = NewWithTemplate("{{ .Payload }}")
	assert.NotNil(tmpl)
	assert.Equal("{{ .Payload }}", tmpl.tmplString)
	assert.Equal(0, len(tmpl.data))
}

func Test_WithData(t *testing.T) {
	assert := assert.New(t)

	tmpl := New().WithTemplate("").WithData("test", true)
	assert.NotNil(tmpl)
	assert.Equal("", tmpl.tmplString)
	assert.Equal(1, len(tmpl.data))
	assert.Equal(true, tmpl.data["test"])
}

func Test_WithRequest(t *testing.T) {
	assert := assert.New(t)

	tmpl := New().WithTemplate("").WithRequest(httptest.NewRequest("GET", "/", nil))
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

	tmpl := New().WithTemplate("").WithPayload(data)
	assert.NotNil(tmpl)
	assert.Equal("", tmpl.tmplString)
	assert.Equal(1, len(tmpl.data))
	assert.JSONEq(`{"test":"test"}`, tmpl.data["Payload"].(string))
}

func Test_Render(t *testing.T) {
	assert := assert.New(t)

	// Test with no template
	_, err := New().Render()
	assert.ErrorIs(err, ErrNoTemplate)

	// Test with basic template
	tmpl := New().WithTemplate("{{ .Payload }}").WithPayload([]byte(`{"test": "test"}`))
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

	tmpl = New().WithTemplate(`
	{
		"customData": {{ toJson .CustomData }},
		"metadata": {
			"testID": "{{ .Request.Header | getHeader "X-Test" }}",
			"deliveryID": "{{ .Request.Header | getHeader "X-Delivery" | default "unknown" }}"
		},
		{{ with $payload := fromJson .Payload }}
		"payload": {
			"foo_exists" : {{ $payload.test.foo | toJson }}
		}
		{{ end }} 
	}
	`).
		WithPayload([]byte(`{"test": {"foo": true}}`)).
		WithRequest(req).
		WithData("CustomData", map[string]string{"foo": "bar"})
	assert.NotNil(tmpl)

	str, err = tmpl.Render()
	assert.Nil(err)
	assert.JSONEq(`{
		"customData": {
			"foo": "bar"
		},
		"metadata": {
			"testID": "test",
			"deliveryID": "unknown"
		},
		"payload": {
			"foo_exists": true
		}
	}`, str)

	// Test with template with template error
	tmpl = New().WithTemplate("{{ .Payload }")
	assert.NotNil(tmpl)
	assert.Equal("{{ .Payload }", tmpl.tmplString)

	str, err = tmpl.Render()
	assert.Error(err)
	assert.Contains(err.Error(), "error in your template: ")
	assert.Equal("", str)

	// Test with template with data error
	tmpl = New().WithTemplate("{{ .Request.Method }}").WithRequest(nil)
	assert.NotNil(tmpl)
	assert.Equal("{{ .Request.Method }}", tmpl.tmplString)

	str, err = tmpl.Render()
	assert.Error(err)
	assert.Contains(err.Error(), "error while filling your template: ")
	assert.Equal("", str)

	// Test with template with invalid format sended to a function
	tmpl = New().WithTemplate(`{{ lookup "test" .Payload }}`).WithPayload([]byte(`{"test": "test"}`))
	assert.NotNil(tmpl)
	assert.Equal(`{{ lookup "test" .Payload }}`, tmpl.tmplString)

	str, err = tmpl.Render()
	assert.Error(err)
	assert.Contains(err.Error(), "template cannot be rendered, check your template")
	assert.Equal("", str)
}

func TestFromContext(t *testing.T) {
	// Test case 1: context value is not a *Formatter
	ctx1 := context.Background()
	_, err1 := FromContext(ctx1)
	assert.Equal(t, ErrNotFoundInContext, err1)

	// Test case 2: context value is a *Formatter
	ctx2 := context.WithValue(context.Background(), formatterCtxKey, &Formatter{})
	formatter, err2 := FromContext(ctx2)
	assert.NotNil(t, formatter)
	assert.Nil(t, err2)
}

func TestToContext(t *testing.T) {
	// Test case 1: context value is nil
	ctx1 := context.Background()
	ctx1 = ToContext(ctx1, nil)
	assert.Nil(t, ctx1.Value(formatterCtxKey))

	// Test case 2: context value is not nil
	ctx2 := context.Background()
	formatter := &Formatter{}
	ctx2 = ToContext(ctx2, formatter)
	assert.Equal(t, formatter, ctx2.Value(formatterCtxKey))
}
