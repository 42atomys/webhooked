package formatting

import (
	"encoding/json"
	"net/http"
	"reflect"
	"text/template"

	"github.com/rs/zerolog/log"
)

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

func dft(dft interface{}, given ...interface{}) interface{} {

	if empty(given) || empty(given[0]) {
		return dft
	}
	return given[0]
}

// empty returns true if the given value has the zero value for its type.
func empty(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	// Basically adapted from text/template.isTrue
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

// coalesce returns the first non-empty value.
func coalesce(v ...interface{}) interface{} {
	for _, val := range v {
		if !empty(val) {
			return val
		}
	}
	return nil
}

// toJson encodes an item into a JSON string
func toJson(v interface{}) string {
	output, err := json.Marshal(v)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal to JSON")
	}
	return string(output)
}

// toPrettyJson encodes an item into a pretty (indented) JSON string
func toPrettyJson(v interface{}) string {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal to JSON")
	}
	return string(output)
}

// ternary returns the first value if the last value is true, otherwise returns the second value.
func ternary(isTrue interface{}, isFalse interface{}, confition bool) interface{} {
	if confition {
		return isTrue
	}

	return isFalse
}

func getHeader(name string, headers *http.Header) string {
	if headers == nil {
		log.Error().Msg("headers are nil. Returning empty string")
		return ""
	}
	return headers.Get(name)
}
