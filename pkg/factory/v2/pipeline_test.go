package factory

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeline(t *testing.T) {
	// f := NewHeaderFactory()

	// err := f.WithOutput("invalid", 1234)
	// assert.EqualError(t, err, "variable invalid is not registered for header")

	// err = f.WithOutput("value", 1234)
	// assert.EqualError(t, err, "invalid type for output value")

	// err = f.WithOutput("value", "test")
	// assert.NoError(t, err)
}

func TestHeaderWithFakeData(t *testing.T) {
	headerName := "Test-Header"
	req := httptest.NewRequest("POST", "/v1alpha1/webhooks/test", nil)
	req.Header.Set(headerName, "testValue")

	f := NewFactory(&headerFactory{})

	f.WithInput("request", req)
	f.WithInput("headerName", headerName)

	f.Run()

	v, ok := GetVar(f.Outputs, "value")
	assert.True(t, ok)
	assert.Equal(t, "testValue", v.Value)
}
