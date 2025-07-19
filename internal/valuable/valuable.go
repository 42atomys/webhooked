// Package valuable provides a flexible way to handle string values that can be retrieved
// from multiple sources, such as direct assignment, environment variables, files,
// or static references.
package valuable

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// Valuable represents a value that can be retrieved in multiple ways.
// It can be a simple value, multiple values, or a reference to an external data source.
type Valuable struct {
	// Value represents a single string value.
	Value *string `json:"value,omitempty"`
	// Values represents multiple string values stored in a slice.
	Values []string `json:"values,omitempty"`
	// ValueFrom represents a reference to an external data source.
	ValueFrom *ValueFromSource `json:"valueFrom,omitempty"`

	// cachedValues caches the computed values to improve performance.
	cachedValues []string
}

// ValueFromSource represents the `valueFrom` field of a configuration entry
// that contains a reference to an external data source (file, environment variable, etc.).
type ValueFromSource struct {
	// StaticRef represents a static value. Can contain a comma-separated list.
	StaticRef *string `json:"staticRef,omitempty"`
	// EnvRef represents a reference to an environment variable.
	EnvRef *string `json:"envRef,omitempty"`
	// FileRef represents a reference to a file.
	FileRef *string `json:"fileRef,omitempty"`
}

// Validate checks the Valuable object and returns an error if any validation fails.
// In the case of EnvRef, the environment variable must exist.
// For FileRef, the file must exist.
func (v *Valuable) Validate() error {
	if v.ValueFrom == nil {
		return nil
	}

	if v.ValueFrom.EnvRef != nil {
		if _, ok := os.LookupEnv(*v.ValueFrom.EnvRef); !ok {
			return fmt.Errorf("environment variable %s not found", *v.ValueFrom.EnvRef)
		}
	}
	if v.ValueFrom.FileRef != nil {
		if _, err := os.Stat(*v.ValueFrom.FileRef); os.IsNotExist(err) {
			return fmt.Errorf("file %s not found", *v.ValueFrom.FileRef)
		}
	}

	return nil
}

// Serialize converts any data into a Valuable and retrieves data from external sources.
// It supports string values.
// @param data is the data to serialize.
// @return the serialized Valuable.
func Serialize(data any) (*Valuable, error) {
	v := &Valuable{}
	switch t := data.(type) {
	case nil:
		return &Valuable{}, nil
	case string:
		v.Value = &t
	case map[string]any:
		// Decode the map into the Valuable struct
		decoderConfig := &mapstructure.DecoderConfig{
			Result:  v,
			TagName: "json",
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				decodeHookMapInterfaceToMapString,
			),
		}
		decoder, err := mapstructure.NewDecoder(decoderConfig)
		if err != nil {
			return nil, err
		}
		if err := decoder.Decode(data); err != nil {
			return nil, fmt.Errorf("unsupported data type %T: %v", data, err)
		}
	default:
		return nil, fmt.Errorf("unsupported data type %T", data)
	}

	// Retrieve data from external sources during serialization
	if err := v.retrieveData(); err != nil {
		return nil, err
	}

	if err := v.Validate(); err != nil {
		return nil, err
	}

	return v, nil
}

// retrieveData fetches data from external sources and caches it.
// This function is called during serialization.
func (v *Valuable) retrieveData() error {
	var computedValues []string

	if len(v.Values) > 0 {
		computedValues = append(computedValues, v.Values...)
	}

	if v.Value != nil && !contains(computedValues, *v.Value) {
		computedValues = append(computedValues, *v.Value)
	}

	if v.ValueFrom != nil {
		if v.ValueFrom.StaticRef != nil {
			computedValues = appendCommaListIfAbsent(computedValues, *v.ValueFrom.StaticRef)
		}

		if v.ValueFrom.EnvRef != nil {
			envValue := os.Getenv(*v.ValueFrom.EnvRef)
			computedValues = appendCommaListIfAbsent(computedValues, envValue)
		}

		if v.ValueFrom.FileRef != nil {
			fileContent, err := os.ReadFile(*v.ValueFrom.FileRef)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", *v.ValueFrom.FileRef, err)
			}
			fileValue := string(fileContent)
			computedValues = append(computedValues, strings.TrimSpace(fileValue))
		}
	}

	v.cachedValues = computedValues
	return nil
}

// decodeHookMapInterfaceToMapString is a decode hook for mapstructure
// that converts map[any]any to map[string]any.
func decodeHookMapInterfaceToMapString(
	f reflect.Type, t reflect.Type, data any,
) (any, error) {
	if f.Kind() != reflect.Map || t.Kind() != reflect.Map {
		return data, nil
	}

	if f.Key().Kind() == reflect.String {
		// No conversion needed
		return data, nil
	}

	mapData, ok := data.(map[any]any)
	if !ok {
		return data, nil
	}

	newMap := make(map[string]any, len(mapData))
	for k, v := range mapData {
		keyStr := fmt.Sprint(k)
		newMap[keyStr] = v
	}
	return newMap, nil
}

// Get returns all cached values of the Valuable as a slice.
// @return the slice of values.
func (v *Valuable) Get() []string {
	return v.cachedValues
}

// First returns the first possible value of the Valuable.
// The order of preference is:
// - Values
// - Value
// - ValueFrom.StaticRef
// - ValueFrom.EnvRef
// - ValueFrom.FileRef
// @return the first value.
func (v *Valuable) First() string {
	if len(v.cachedValues) == 0 {
		return ""
	}
	return v.cachedValues[0]
}

// String returns the string representation of the first value.
func (v Valuable) String() string {
	return v.First()
}

// Contains returns true if the Valuable contains the given value.
// @param element is the value to check.
// @return true if the Valuable contains the given value.
func (v *Valuable) Contains(element string) bool {
	return contains(v.cachedValues, element)
}

// contains checks if a slice contains a specific string.
func contains(slice []string, element string) bool {
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}

// appendCommaListIfAbsent accepts a comma-separated list of strings to append
// to the slice only if the element is absent.
func appendCommaListIfAbsent(slice []string, commaList string) []string {
	items := strings.Split(commaList, ",")
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if !contains(slice, s) {
			slice = append(slice, s)
		}
	}
	return slice
}
