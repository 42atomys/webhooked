package formatting

import (
	"encoding/json"
	"net/http"
	"reflect"
	"text/template"

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
		"ternary":      ternary,

		// Headers manipulation functions
		"getHeader": getHeader,
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
		return !g.Bool()
	case reflect.Complex64, reflect.Complex128:
		return g.Complex() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return g.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return g.Float() == 0
	case reflect.Struct:
		return false
	}
}

// coalesce returns the first value not empty in the given list.
// If all values are empty, it returns nil.
func coalesce(v ...interface{}) interface{} {
	for _, val := range v {
		if !empty(val) {
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

// ternary returns `isTrue` if `condition` is true, otherwise returns `isFalse`.
func ternary(isTrue interface{}, isFalse interface{}, confition bool) interface{} {
	if confition {
		return isTrue
	}

	return isFalse
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
