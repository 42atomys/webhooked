package formatting

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_funcMap(t *testing.T) {
	assert := assert.New(t)

	funcMap := funcMap()
	assert.Contains(funcMap, "default")
	assert.NotContains(funcMap, "dft")
	assert.Contains(funcMap, "empty")
	assert.Contains(funcMap, "coalesce")
	assert.Contains(funcMap, "toJson")
	assert.Contains(funcMap, "toPrettyJson")
	assert.Contains(funcMap, "ternary")
	assert.Contains(funcMap, "getHeader")
}

func Test_dft(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("test", dft("default", "test"))
	assert.Equal("default", dft("default", nil))
	assert.Equal("default", dft("default", ""))
}

func Test_empty(t *testing.T) {
	assert := assert.New(t)

	assert.True(empty(""))
	assert.True(empty(nil))
	assert.False(empty("test"))
	assert.False(empty(true))
	assert.True(empty(false))
	assert.True(empty(0 + 0i))
	assert.False(empty(2 + 4i))
	assert.True(empty([]int{}))
	assert.False(empty([]int{1}))
	assert.True(empty(map[string]string{}))
	assert.False(empty(map[string]string{"test": "test"}))
	assert.True(empty(map[string]interface{}{}))
	assert.False(empty(map[string]interface{}{"test": "test"}))
	assert.True(empty(0))
	assert.False(empty(-1))
	assert.False(empty(1))
	assert.True(empty(uint32(0)))
	assert.False(empty(uint32(1)))
	assert.True(empty(float64(0.0)))
	assert.False(empty(float64(1.0)))
	assert.False(empty(struct{}{}))
	assert.False(empty(struct{ Test string }{Test: "test"}))

	ptr := &struct{ Test string }{Test: "test"}
	assert.False(empty(ptr))
}

func Test_coalesce(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("test", coalesce("test", "default"))
	assert.Equal("default", coalesce("", "default"))
	assert.Equal("default", coalesce(nil, "default"))
	assert.Equal(nil, coalesce(nil, nil))
}

func Test_toJson(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("{\"test\":\"test\"}", toJson(map[string]string{"test": "test"}))
	assert.Equal("{\"test\":\"test\"}", toJson(map[string]interface{}{"test": "test"}))
	assert.Equal("null", toJson(nil))
	assert.Equal("", toJson(map[string]interface{}{"test": func() {}}))
}

func Test_toPrettyJson(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("{\n  \"test\": \"test\"\n}", toPrettyJson(map[string]string{"test": "test"}))
	assert.Equal("{\n  \"test\": \"test\"\n}", toPrettyJson(map[string]interface{}{"test": "test"}))
	assert.Equal("null", toPrettyJson(nil))
	assert.Equal("", toPrettyJson(map[string]interface{}{"test": func() {}}))
}

func Test_ternary(t *testing.T) {
	assert := assert.New(t)

	header := httptest.NewRecorder().Header()

	header.Set("X-Test", "test")
	assert.Equal("test", getHeader("X-Test", &header))
	assert.Equal("", getHeader("X-Undefined", &header))
	assert.Equal("", getHeader("", &header))
	assert.Equal("", getHeader("", nil))
}

func Test_getHeader(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(true, ternary(true, false, true))
	assert.Equal(false, ternary(true, false, false))
	assert.Equal("true string", ternary("true string", "false string", true))
	assert.Equal("false string", ternary("true string", "false string", false))
	assert.Equal(nil, ternary(nil, nil, false))
}
