package formatting

import (
	"bytes"
	"fmt"
	"math"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsNumber(t *testing.T) {
	// Test with nil value
	assert.False(t, isNumber(nil))

	// Test with invalid value
	assert.False(t, isNumber(math.NaN()))
	assert.False(t, isNumber(math.Inf(1)))
	assert.False(t, isNumber(math.Inf(-1)))
	assert.False(t, isNumber(complex(1, 2)))

	// Test with integer values
	assert.True(t, isNumber(int(42)))
	assert.True(t, isNumber(int8(42)))
	assert.True(t, isNumber(int16(42)))
	assert.True(t, isNumber(int32(42)))
	assert.True(t, isNumber(int64(42)))
	assert.True(t, isNumber(uint(42)))
	assert.True(t, isNumber(uint8(42)))
	assert.True(t, isNumber(uint16(42)))
	assert.True(t, isNumber(uint32(42)))
	assert.True(t, isNumber(uint64(42)))
	assert.False(t, isNumber(uintptr(42)))

	// Test with floating-point values
	assert.True(t, isNumber(float32(3.14)))
	assert.True(t, isNumber(float64(3.14)))

}

type customStringer struct {
	str string
}

func (s customStringer) String() string {
	return s.str
}

func TestIsString(t *testing.T) {
	// Test with nil value
	assert.False(t, isString(nil))

	// Test with empty value
	assert.False(t, isString(""))
	assert.False(t, isString([]byte{}))
	assert.False(t, isString(struct{}{}))

	// Test with non-empty value
	assert.True(t, isString("test"))
	assert.True(t, isString([]byte("test")))
	assert.True(t, isString(fmt.Sprintf("%v", 42)))
	assert.True(t, isString(customStringer{}))
	assert.True(t, isString(time.Now()))
	assert.False(t, isString(42))
	assert.False(t, isString(3.14))
	assert.False(t, isString([]int{1, 2, 3}))
	assert.False(t, isString(httptest.NewRecorder()))
	assert.False(t, isString(struct{ String string }{String: "test"}))
	assert.False(t, isString(map[string]string{"foo": "bar"}))
}

func TestIsBool(t *testing.T) {
	// Test with a bool value
	assert.True(t, isBool(true))
	assert.True(t, isBool(false))

	// Test with a string value
	assert.True(t, isBool("true"))
	assert.True(t, isBool("false"))
	assert.True(t, isBool("TRUE"))
	assert.True(t, isBool("FALSE"))
	assert.False(t, isBool("foo"))
	assert.False(t, isBool(""))

	// Test with a []byte value
	assert.True(t, isBool([]byte("true")))
	assert.True(t, isBool([]byte("false")))
	assert.True(t, isBool([]byte("TRUE")))
	assert.True(t, isBool([]byte("FALSE")))
	assert.False(t, isBool([]byte("foo")))
	assert.False(t, isBool([]byte("")))

	// Test with a fmt.Stringer value
	assert.True(t, isBool(fmt.Sprintf("%v", true)))
	assert.True(t, isBool(fmt.Sprintf("%v", false)))
	assert.False(t, isBool(fmt.Sprintf("%v", 42)))

	// Test with other types
	assert.False(t, isBool(nil))
	assert.False(t, isBool(42))
	assert.False(t, isBool(3.14))
	assert.False(t, isBool([]int{1, 2, 3}))
	assert.False(t, isBool(map[string]string{"foo": "bar"}))
	assert.False(t, isBool(struct{ Foo string }{Foo: "bar"}))
}

func TestIsNull(t *testing.T) {
	// Test with nil value
	assert.True(t, isNull(nil))

	// Test with empty value
	assert.True(t, isNull(""))
	assert.True(t, isNull([]int{}))
	assert.True(t, isNull(map[string]string{}))
	assert.True(t, isNull(struct{}{}))

	// Test with non-empty value
	assert.False(t, isNull("test"))
	assert.False(t, isNull(42))
	assert.False(t, isNull(3.14))
	assert.False(t, isNull([]int{1, 2, 3}))
	assert.False(t, isNull(map[string]string{"foo": "bar"}))
	assert.False(t, isNull(struct{ Foo string }{Foo: "bar"}))
	assert.False(t, isNull(time.Now()))
	assert.False(t, isNull(httptest.NewRecorder()))
}

func TestToString(t *testing.T) {
	// Test with nil value
	assert.Equal(t, "", toString(nil))

	// Test with invalid value
	buf := new(bytes.Buffer)
	assert.Equal(t, "", toString(buf))

	// Test with string value
	assert.Equal(t, "test", toString("test"))
	assert.Equal(t, "test", toString([]byte("test")))
	assert.Equal(t, "42", toString(fmt.Sprintf("%v", 42)))
	assert.Equal(t, "", toString(struct{ String string }{String: "test"}))
	assert.Equal(t, "", toString(struct{}{}))

	// Test with fmt.Stringer value
	assert.Equal(t, "test", toString(customStringer{str: "test"}))
	assert.Equal(t, "", toString(customStringer{}))

	// Test with other types
	assert.Equal(t, "42", toString(42))
	assert.Equal(t, "42", toString(uint(42)))
	assert.Equal(t, "3.14", toString(3.14))
	assert.Equal(t, "true", toString(true))
	assert.Equal(t, "false", toString(false))
	assert.Equal(t, "2009-11-10 23:00:00 +0000 UTC", toString(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)))
}

func TestToInt(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0, toInt(nil))

	// Test with invalid value
	assert.Equal(t, 0, toInt("test"))
	assert.Equal(t, 0, toInt([]byte("test")))
	assert.Equal(t, 0, toInt(struct{ Int int }{Int: 42}))
	assert.Equal(t, 0, toInt(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 42, toInt(42))
	assert.Equal(t, -42, toInt("-42"))
	assert.Equal(t, 0, toInt("0"))
	assert.Equal(t, 123456789, toInt("123456789"))
}

func TestToFloat(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, toFloat(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, toFloat("test"))
	assert.Equal(t, 0.0, toFloat([]byte("test")))
	assert.Equal(t, 0.0, toFloat(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, toFloat(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 42.0, toFloat(42))
	assert.Equal(t, -42.0, toFloat("-42"))
	assert.Equal(t, 0.0, toFloat("0"))
	assert.Equal(t, 123456789.0, toFloat("123456789"))
	assert.Equal(t, 3.14, toFloat(3.14))
	assert.Equal(t, 2.71828, toFloat("2.71828"))
}

func TestToBool(t *testing.T) {
	// Test with nil value
	assert.False(t, toBool(nil))

	// Test with invalid value
	assert.False(t, toBool("test"))
	assert.False(t, toBool([]byte("test")))
	assert.False(t, toBool(struct{ Bool bool }{Bool: true}))
	assert.False(t, toBool(new(bytes.Buffer)))

	// Test with valid value
	assert.True(t, toBool(true))
	assert.True(t, toBool("true"))
	assert.True(t, toBool("1"))
	assert.False(t, toBool(false))
	assert.False(t, toBool("false"))
	assert.False(t, toBool("0"))
}
