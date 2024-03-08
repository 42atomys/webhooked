package formatting

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
)

// funcMap is the map of functions that can be used in templates.
// The key is the name of the function and the value is the function itself.
// This is required for the template.New() function to parse the function.
func funcMap() template.FuncMap {
	return template.FuncMap{
		// Core functions
		"default":      dft,
		"empty":        empty,
		"coalesce":     coalesce,
		"toJson":       toJson,
		"toPrettyJson": toPrettyJson,
		"fromJson":     fromJson,
		"ternary":      ternary,
		"lookup":       lookup,

		// Headers manipulation functions
		"getHeader": getHeader,

		// Time manipulation functions
		"formatTime": formatTime,
		"parseTime":  parseTime,

		// Casting functions
		"toString": toString,
		"toInt":    toInt,
		"toFloat":  toFloat,
		"toBool":   toBool,

		// Is functions
		"isNumber": isNumber,
		"isString": isString,
		"isBool":   isBool,
		"isNull":   isNull,

		// Math functions
		"add":  mathAdd,
		"sub":  mathSub,
		"mul":  mathMul,
		"div":  mathDiv,
		"mod":  mathMod,
		"pow":  mathPow,
		"max":  mathMax,
		"min":  mathMin,
		"sqrt": mathSqrt,
	}
}

// dft returns the default value if the given value is empty.
// If the given value is not empty, it is returned as is.
func dft(dft interface{}, given ...interface{}) interface{} {

	if empty(given) || empty(given[0]) {
		return dft
	}
	return given[0]
}

// empty returns true if the given value is empty.
// It supports any type.
func empty(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	switch g.Kind() {
	default:
		return g.IsNil()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return g.Len() == 0
	case reflect.Bool:
		return !g.IsValid()
	case reflect.Complex64, reflect.Complex128:
		return g.Complex() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return g.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return g.Float() == 0
	case reflect.Struct:
		return g.NumField() == 0
	}
}

// coalesce returns the first value not empty in the given list.
// If all values are empty, it returns nil.
func coalesce(v ...interface{}) interface{} {
	for _, val := range v {
		if !isNull(val) {
			return val
		}
	}
	return nil
}

// toJson returns the given value as a JSON string.
// If the given value is nil, it returns an empty string.
func toJson(v interface{}) string {
	output, err := json.Marshal(v)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal to JSON")
	}
	return string(output)
}

// toPrettyJson returns the given value as a pretty JSON string indented with
// 2 spaces. If the given value is nil, it returns an empty string.
func toPrettyJson(v interface{}) string {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal to JSON")
	}
	return string(output)
}

// fromJson returns the given JSON string as a map[string]interface{}.
// If the given value is nil, it returns an empty map.
func fromJson(v interface{}) map[string]interface{} {
	if isNull(v) {
		return map[string]interface{}{}
	}

	if v, ok := v.(map[string]interface{}); ok {
		return v
	}

	var output = map[string]interface{}{}
	var err error
	if bytes, ok := v.([]byte); ok {
		err = json.Unmarshal(bytes, &output)
	} else {
		err = json.Unmarshal([]byte(v.(string)), &output)
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal JSON")
	}
	return output
}

// ternary returns `isTrue` if `condition` is true, otherwise returns `isFalse`.
func ternary(isTrue interface{}, isFalse interface{}, condition bool) interface{} {
	if condition {
		return isTrue
	}

	return isFalse
}

// lookup recursively navigates through nested data structures based on a dot-separated path.
func lookup(path string, data interface{}) interface{} {
	keys := strings.Split(path, ".")

	if path == "" {
		return data
	}

	// Navigate through the data for each key.
	current := data
	for _, key := range keys {
		switch val := current.(type) {
		case map[string]interface{}:
			// If the current value is a map and the key exists, proceed to the next level.
			if next, ok := val[key]; ok {
				current = next
			} else {
				// Key not found
				log.Logger.Warn().Str("path", path).Msg("Key are not found on the object")
				return nil
			}
		default:
			// If the current type is not a map or we've reached a non-navigable point
			return nil
		}
	}

	// If the final value is a string, return it; otherwise
	return current
}

// getHeader returns the value of the given header. If the header is not found,
// it returns an empty string.
func getHeader(name string, headers *http.Header) string {
	if headers == nil {
		log.Error().Msg("headers are nil. Returning empty string")
		return ""
	}
	return headers.Get(name)
}

// formatTime returns the given time formatted with the given layout.
// If the given time is invalid, it returns an empty string.
func formatTime(t interface{}, fromLayout, tolayout string) string {
	if isNull(t) {
		log.Error().Msg("time is nil. Returning empty string")
		return ""
	}

	if tolayout == "" {
		tolayout = time.RFC3339
	}

	parsedTime := parseTime(t, fromLayout)
	if parsedTime.IsZero() {
		log.Error().Msgf("Failed to parse time [%v] with layout [%s]", t, fromLayout)
		return ""
	}

	return parsedTime.Format(tolayout)
}

// parseTime returns the given time parsed with the given layout.
// If the given time is invalid, it returns an time.Time{}.
func parseTime(t interface{}, layout string) time.Time {
	if isNull(t) {
		return time.Time{}
	}

	var parsedTime time.Time
	var err error
	switch reflect.ValueOf(t).Kind() {
	default:
		t, ok := t.(time.Time)
		if ok {
			parsedTime = t
		} else {
			parsedTime = time.Time{}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsedTime = time.Unix(int64(toInt(t)), 0)
	case reflect.String:
		parsedTime, err = time.Parse(layout, toString(t))
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to parse time")
		return time.Time{}
	}

	return parsedTime
}

// isNumber returns true if the given value is a number, otherwise returns false.
func isNumber(n interface{}) bool {
	if isNull(n) {
		return false
	}

	g := reflect.ValueOf(n)
	switch g.Kind() {
	default:
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return !math.IsNaN(g.Float()) && !(math.IsInf(g.Float(), 1) || math.IsInf(g.Float(), -1))
	case reflect.Uintptr:
		return false
	}
}

// isString returns true if the given value is a string, otherwise returns false.
func isString(n interface{}) bool {
	if isNull(n) {
		return false
	}

	switch n.(type) {
	default:
		if _, ok := n.(fmt.Stringer); ok {
			return true
		}
		return false
	case string, []byte:
		return true
	}
}

// isBool returns true if the given value is a bool, otherwise returns false.
func isBool(n interface{}) bool {
	if isNull(n) {
		return false
	}

	switch n.(type) {
	default:
		return false
	case string, []byte, fmt.Stringer:
		_, err := strconv.ParseBool(toString(n))
		return err == nil
	case bool:
		return true
	}
}

// isNull returns true if the given value is nil or empty, otherwise returns false.
func isNull(n interface{}) bool {
	if n == nil || empty(n) {
		return true
	}

	return false
}

// toString returns the given value as a string.
// If the given value is nil, it returns an empty string.
func toString(n interface{}) string {
	if isNull(n) {
		return ""
	}

	switch n := n.(type) {
	default:
		g := reflect.ValueOf(n)
		switch g.Kind() {
		default:
			return ""
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(g.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.FormatUint(g.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			return strconv.FormatFloat(g.Float(), 'f', -1, 64)
		case reflect.Bool:
			return strconv.FormatBool(g.Bool())
		}
	case string, []byte:
		return fmt.Sprintf("%s", n)
	case fmt.Stringer:
		return n.String()
	}
}

// toInt returns the given value as an int.
// If the given value is nil, it returns 0.
func toInt(n interface{}) int {
	if isNull(n) {
		return 0
	}

	i, err := strconv.Atoi(toString(n))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to convert [%v] to int", n)
		return 0
	}

	return i
}

// toFloat returns the given value as a float.
// If the given value is nil, it returns 0.
func toFloat(n interface{}) float64 {
	if isNull(n) {
		return 0
	}

	f, err := strconv.ParseFloat(toString(n), 64)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to convert [%v] to float", n)
		return 0
	}

	return f
}

// toBool returns the given value as a bool.
// If the given value is nil, it returns false.
func toBool(n interface{}) bool {
	if isNull(n) {
		return false
	}

	b, err := strconv.ParseBool(toString(n))
	if err != nil {
		log.Error().Err(err).Msgf("Failed to convert [%v] to bool", n)
		return false
	}

	return b
}

// mathAdd returns the sum of the given numbers.
// If any of the given numbers is not a number, it returns 0.
func mathAdd(numbers ...interface{}) float64 {
	var sum float64
	for _, n := range numbers {
		sum += toFloat(n)
	}
	return sum
}

// mathSub returns the difference of the given numbers.
// If any of the given numbers is not a number, it returns 0.
func mathSub(numbers ...interface{}) float64 {
	var diff float64
	for i, n := range numbers {
		f := toFloat(n)

		if i == 0 {
			diff = f
			continue
		}

		diff -= f
	}
	return diff
}

// mathMul returns the product of the given numbers.
// If any of the given numbers is not a number, it returns 0.
func mathMul(numbers ...interface{}) float64 {
	var product float64
	for _, n := range numbers {
		p := toFloat(n)

		if product == 0 {
			product = p
			continue
		}

		product *= p
	}
	return product
}

// mathDiv returns the quotient of the given numbers.
// If any of the given numbers is not a number, it returns 0.
func mathDiv(numbers ...interface{}) float64 {
	var quotient float64
	for i, n := range numbers {
		d := toFloat(n)

		if i == 0 {
			quotient = d
			continue
		}

		quotient /= d
	}
	return quotient
}

// mathMod returns the remainder of the given numbers.
// If any of the given numbers is not a number, it returns 0.
func mathMod(numbers ...interface{}) float64 {
	var remainder float64
	for i, n := range numbers {
		m := toFloat(n)

		if i == 0 {
			remainder = m
			continue
		}

		remainder = math.Mod(remainder, m)
	}
	return remainder
}

// mathPow returns the power of the given numbers.
// If any of the given numbers is not a number, it returns 0.
func mathPow(numbers ...interface{}) float64 {
	var power float64
	for i, n := range numbers {
		p := toFloat(n)

		if i == 0 {
			power = p
			continue
		}

		power = math.Pow(power, p)
	}
	return power
}

// mathSqrt returns the square root of the given number.
// If the given number is not a number, it returns 0.
func mathSqrt(number interface{}) float64 {
	return math.Sqrt(toFloat(number))
}

// mathMin returns the minimum of the given numbers.
// If any of the given numbers is not a number, it returns 0.
func mathMin(numbers ...interface{}) float64 {
	var min float64
	for i, n := range numbers {
		num := toFloat(n)

		if i == 0 {
			min = num
			continue
		}

		if num < min {
			min = num
		}
	}
	return min
}

// mathMax returns the maximum of the given numbers.
// If any of the given numbers is not a number, it returns 0.
func mathMax(numbers ...interface{}) float64 {
	var max float64
	for i, n := range numbers {
		num := toFloat(n)

		if i == 0 {
			max = num
			continue
		}

		if num > max {
			max = num
		}
	}
	return max
}
