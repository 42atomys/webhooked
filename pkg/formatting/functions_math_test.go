package formatting

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMathAdd(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathAdd(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathAdd("test"))
	assert.Equal(t, 0.0, mathAdd([]byte("test")))
	assert.Equal(t, 0.0, mathAdd(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathAdd(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 42.0, mathAdd(42))
	assert.Equal(t, 0.0, mathAdd())
	assert.Equal(t, 6.0, mathAdd(1, 2, 3))
	assert.Equal(t, 10.0, mathAdd(1, 2, "3", 4))
	assert.Equal(t, 3.14, mathAdd(3.14))
	assert.Equal(t, 5.0, mathAdd(2, 3.0))
}

func TestMathSub(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathSub(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathSub("test"))
	assert.Equal(t, 0.0, mathSub([]byte("test")))
	assert.Equal(t, 0.0, mathSub(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathSub(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 42.0, mathSub(42))
	assert.Equal(t, 0.0, mathSub())
	assert.Equal(t, -4.0, mathSub(1, 2, 3))
	assert.Equal(t, -8.0, mathSub(1, 2, "3", 4))
	assert.Equal(t, 3.14, mathSub(3.14))
	assert.Equal(t, -1.0, mathSub(2, 3.0))
}

func TestMathMul(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathMul(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathMul("test"))
	assert.Equal(t, 0.0, mathMul([]byte("test")))
	assert.Equal(t, 0.0, mathMul(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathMul(new(bytes.Buffer)))
	assert.Equal(t, 0.0, mathMul())

	// Test with valid value
	assert.Equal(t, 42.0, mathMul(42))
	assert.Equal(t, 6.0, mathMul(1, 2, 3))
	assert.Equal(t, 24.0, mathMul(1, 2, "3", 4))
	assert.Equal(t, 3.14, mathMul(3.14))
	assert.Equal(t, 6.0, mathMul(2, 3.0))
}

func TestMathDiv(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathDiv(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathDiv("test"))
	assert.Equal(t, 0.0, mathDiv([]byte("test")))
	assert.Equal(t, 0.0, mathDiv(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathDiv(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 42.0, mathDiv(42))
	assert.Equal(t, 0.0, mathDiv())
	assert.Equal(t, 0.16666666666666666, mathDiv(1, 2, 3))
	assert.Equal(t, 0.041666666666666664, mathDiv(1, 2, "3", 4))
	assert.Equal(t, 3.14, mathDiv(3.14))
	assert.Equal(t, 0.6666666666666666, mathDiv(2, 3.0))
}

func TestMathMod(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathMod(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathMod("test"))
	assert.Equal(t, 0.0, mathMod([]byte("test")))
	assert.Equal(t, 0.0, mathMod(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathMod(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 42.0, mathMod(42))
	assert.Equal(t, 0.0, mathMod())
	assert.Equal(t, 1.0, mathMod(10, 3, 2))
	assert.Equal(t, 0.0, mathMod(10, 2))
	assert.Equal(t, 1.0, mathMod(10, 3))
	assert.Equal(t, 0.0, mathMod(10, 5))
	assert.Equal(t, 0.5, mathMod(10.5, 2))
}

func TestMathPow(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathPow(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathPow("test"))
	assert.Equal(t, 0.0, mathPow([]byte("test")))
	assert.Equal(t, 0.0, mathPow(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathPow(new(bytes.Buffer)))
	assert.Equal(t, 0.0, mathPow())

	// Test with valid value
	assert.Equal(t, 2.0, mathPow(2))
	assert.Equal(t, 8.0, mathPow(2, 3))
	assert.Equal(t, 64.0, mathPow(2, 3, 2))
	assert.Equal(t, 1.0, mathPow(2, 0))
	assert.Equal(t, 0.25, mathPow(2, -2))
	assert.Equal(t, 27.0, mathPow(3, "3"))
	assert.Equal(t, 4.0, mathPow(2, 2.0))
}

func TestMathSqrt(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathSqrt(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathSqrt("test"))
	assert.Equal(t, 0.0, mathSqrt([]byte("test")))
	assert.Equal(t, 0.0, mathSqrt(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathSqrt(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 2.0, mathSqrt(4))
	assert.Equal(t, 3.0, mathSqrt(9))
	assert.Equal(t, 0.0, mathSqrt(0))
	assert.Equal(t, math.Sqrt(2), mathSqrt(2))
	assert.Equal(t, math.Sqrt(0.5), mathSqrt(0.5))
}

func TestMathMin(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathMin(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathMin("test"))
	assert.Equal(t, 0.0, mathMin([]byte("test")))
	assert.Equal(t, 0.0, mathMin(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathMin(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 1.0, mathMin(1))
	assert.Equal(t, 2.0, mathMin(2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12))
	assert.Equal(t, -1.0, mathMin(-1, 0, 1))
	assert.Equal(t, 0.0, mathMin(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
	assert.Equal(t, 0.5, mathMin(1, 0.5, 2))
	assert.Equal(t, -2.0, mathMin(2, -2, 0))
}

func TestMathMax(t *testing.T) {
	// Test with nil value
	assert.Equal(t, 0.0, mathMax(nil))

	// Test with invalid value
	assert.Equal(t, 0.0, mathMax("test"))
	assert.Equal(t, 0.0, mathMax([]byte("test")))
	assert.Equal(t, 0.0, mathMax(struct{ Float float64 }{Float: 42}))
	assert.Equal(t, 0.0, mathMax(new(bytes.Buffer)))

	// Test with valid value
	assert.Equal(t, 1.0, mathMax(1))
	assert.Equal(t, 12.0, mathMax(2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12))
	assert.Equal(t, 1.0, mathMax(-1, 0, 1))
	assert.Equal(t, 0.0, mathMax(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
	assert.Equal(t, 2.0, mathMax(1, 0.5, 2))
	assert.Equal(t, 2.0, mathMax(2, -2, 0))
}
