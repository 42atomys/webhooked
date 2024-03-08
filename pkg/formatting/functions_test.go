package formatting

import (
	"net/http/httptest"
	"testing"
	"time"

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
	assert.False(empty(false))
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
	assert.True(empty(struct{}{}))
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

func Test_fromJson(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(map[string]interface{}{"test": "test"}, fromJson("{\"test\":\"test\"}"))
	assert.Equal(map[string]interface{}{"test": map[string]interface{}{"foo": true}}, fromJson("{\"test\":{\"foo\":true}}"))
	assert.Equal(map[string]interface{}{}, fromJson(nil))
	assert.Equal(map[string]interface{}{"test": 1}, fromJson(map[string]interface{}{"test": 1}))
	assert.Equal(map[string]interface{}{}, fromJson(""))
	assert.Equal(map[string]interface{}{"test": "test"}, fromJson([]byte("{\"test\":\"test\"}")))
	assert.Equal(map[string]interface{}{}, fromJson([]byte("\\\\")))

	var result = fromJson("{\"test\":\"test\"}")
	assert.Equal(result["test"], "test")
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

func TestLookup(t *testing.T) {
	// Initialize the assert helper
	assert := assert.New(t)

	// Example of nested data structure for testing
	testData := map[string]interface{}{
		"user": map[string]interface{}{
			"details": map[string]interface{}{
				"name": "John Doe",
				"age":  30,
			},
			"email": "john.doe@example.com",
		},
		"empty": map[string]interface{}{},
	}

	// Test cases
	tests := []struct {
		path     string
		data     interface{}
		expected interface{}
	}{
		// Test successful lookups
		{"user.details.name", testData, "John Doe"},
		{"user.email", testData, "john.doe@example.com"},
		// Test unsuccessful lookups
		{"user.details.phone", testData, nil},
		{"user.location.city", testData, nil},
		// Test edge cases
		{"", testData, testData},
		{"user..name", testData, nil},
		{"nonexistent", testData, nil},
		// Test with non-map data
		{"user", []interface{}{}, nil},
	}

	// Run test cases
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			result := lookup(test.path, test.data)
			assert.Equal(test.expected, result, "Lookup should return the expected value.")
		})
	}
}

func Test_getHeader(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(true, ternary(true, false, true))
	assert.Equal(false, ternary(true, false, false))
	assert.Equal("true string", ternary("true string", "false string", true))
	assert.Equal("false string", ternary("true string", "false string", false))
	assert.Equal(nil, ternary(nil, nil, false))
}

func Test_formatTime(t *testing.T) {
	assert := assert.New(t)

	teaTime := parseTime("2023-01-01T08:42:00Z", time.RFC3339)
	assert.Equal("Sun Jan  1 08:42:00 UTC 2023", formatTime(teaTime, time.RFC3339, time.UnixDate))

	teaTime = parseTime("Mon Jan 01 08:42:00 UTC 2023", time.UnixDate)
	assert.Equal("2023-01-01T08:42:00Z", formatTime(teaTime, time.UnixDate, time.RFC3339))

	// from unix
	teaTime = parseTime("2023-01-01T08:42:00Z", time.RFC3339)
	assert.Equal("Sun Jan  1 08:42:00 UTC 2023", formatTime(teaTime.Unix(), "", time.UnixDate))

	assert.Equal("", formatTime("INVALID_TIME", "", ""))
	assert.Equal("", formatTime(nil, "", ""))
}

func TestParseTime(t *testing.T) {
	// Test with nil value
	assert.Equal(t, time.Time{}, parseTime(nil, ""))
	// Test with invalid value
	assert.Equal(t, time.Time{}, parseTime("test", ""))
	assert.Equal(t, time.Time{}, parseTime(true, ""))
	assert.Equal(t, time.Time{}, parseTime([]byte("test"), ""))
	assert.Equal(t, time.Time{}, parseTime(struct{ Time time.Time }{Time: time.Now()}, ""))
	assert.Equal(t, time.Time{}, parseTime(httptest.NewRecorder(), ""))
	assert.Equal(t, time.Time{}, parseTime("INVALID_TIME", ""))
	assert.Equal(t, time.Time{}, parseTime("", ""))
	assert.Equal(t, time.Time{}, parseTime("", "INVALID_LAYOUT"))

	// Test with valid value
	teaTime := time.Date(2023, 1, 1, 8, 42, 0, 0, time.UTC)
	assert.Equal(t, teaTime, parseTime("2023-01-01T08:42:00Z", time.RFC3339))
	assert.Equal(t, teaTime, parseTime("Mon Jan 01 08:42:00 UTC 2023", time.UnixDate))
	assert.Equal(t, teaTime, parseTime("Monday, 01-Jan-23 08:42:00 UTC", time.RFC850))
	assert.Equal(t, teaTime, parseTime("2023/01/01 08h42m00", "2006/01/02 15h04m05"))
	teaTime = time.Date(2023, 1, 1, 8, 42, 0, 0, time.Local)
	assert.Equal(t, teaTime, parseTime(teaTime.Unix(), ""))

	assert.Equal(t, time.Unix(1234567890, 0), parseTime(int64(1234567890), ""))
	assert.Equal(t, time.Time{}, parseTime(int32(0), ""))
	assert.Equal(t, time.Time{}, parseTime(int16(0), ""))
	assert.Equal(t, time.Time{}, parseTime(int8(0), ""))
	assert.Equal(t, time.Time{}, parseTime(int(0), ""))
	assert.Equal(t, time.Time{}, parseTime(uint(0), ""))
	assert.Equal(t, time.Time{}, parseTime(uint32(0), ""))
	assert.Equal(t, time.Time{}, parseTime(uint64(0), ""))
	assert.Equal(t, time.Time{}, parseTime(float32(0), ""))
	assert.Equal(t, time.Time{}, parseTime(float64(0), ""))
	assert.Equal(t, time.Time{}, parseTime("", ""))
	assert.Equal(t, time.Time{}, parseTime("invalid", ""))
	assert.Equal(t, time.Time{}, parseTime("2006-01-02 15:04:05", ""))
	assert.Equal(t, time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC), parseTime("2022-12-31", "2006-01-02"))
	assert.Equal(t, time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC), parseTime("2022-12-31 23:59:59", "2006-01-02 15:04:05"))
}
