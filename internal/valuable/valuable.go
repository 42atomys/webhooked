package valuable

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// Valuable represent value who it is possible to retrieve the data
// in multiple ways. From a simple value without nesting,
// or from a deep data source.
type Valuable struct {
	// Value represents the `value` field of a configuration entry that
	// contains only one value
	Value *string `json:"value,omitempty"`
	// Values represents the `value` field of a configuration entry that
	// contains multiple values stored in a list
	Values []string `json:"values,omitempty"`
	// ValueFrom represents the `valueFrom` field of a configuration entry
	// that contains a reference to a data source
	ValueFrom *ValueFromSource `json:"valueFrom,omitempty"`
}

// ValueFromSource represents the `valueFrom` field of a configuration entry
// that contains a reference to a data source (file, env, etc.)
type ValueFromSource struct {
	// StaticRef represents the `staticRef` field of a configuration entry
	// that contains a static value. Can contain a comma separated list
	StaticRef *string `json:"staticRef,omitempty"`
	// EnvRef represents the `envRef` field of a configuration entry
	// that contains a reference to an environment variable
	EnvRef *string `json:"envRef,omitempty"`
}

// Validate validates the Valuable object and returns an error if any
// validation fails. In case of envRef, the env variable must exist.
func (v *Valuable) Validate() error {
	if v.ValueFrom != nil && v.ValueFrom.EnvRef != nil {
		if _, ok := os.LookupEnv(*v.ValueFrom.EnvRef); !ok {
			return fmt.Errorf("environment variable %s not found", *v.ValueFrom.EnvRef)
		}
	}

	return nil
}

// SerializeValuable serialize anything to a Valuable
// @param data is the data to serialize
// @return the serialized Valuable
func SerializeValuable(data interface{}) (*Valuable, error) {
	var v *Valuable = &Valuable{}
	switch t := data.(type) {
	case string:
		v.Value = &t
	case int, float32, float64, bool:
		str := fmt.Sprint(t)
		v.Value = &str
	case nil:
		return &Valuable{}, nil
	case map[interface{}]interface{}:
		var val *Valuable
		if err := mapstructure.Decode(data, &val); err != nil {
			return nil, err
		}
		v = val
	default:
		valuable := Valuable{}
		if err := mapstructure.Decode(data, &valuable); err != nil {
			return nil, fmt.Errorf("unimplemented valuable type %s", reflect.TypeOf(data).String())
		}
		v = &valuable
	}

	if err := v.Validate(); err != nil {
		return nil, err
	}

	return v, nil
}

// Get returns all values of the Valuable as a slice
// @return the slice of values
func (v *Valuable) Get() []string {
	var computedValues []string

	computedValues = append(computedValues, v.Values...)

	if v.Value != nil && !contains(computedValues, *v.Value) {
		computedValues = append(computedValues, *v.Value)
	}

	if v.ValueFrom == nil {
		return computedValues
	}

	if v.ValueFrom.StaticRef != nil && !contains(computedValues, *v.ValueFrom.StaticRef) {
		computedValues = appendCommaListIfAbsent(computedValues, *v.ValueFrom.StaticRef)
	}

	if v.ValueFrom.EnvRef != nil {
		computedValues = appendCommaListIfAbsent(computedValues, os.Getenv(*v.ValueFrom.EnvRef))
	}

	return computedValues
}

// First returns the first value of the Valuable possible values
// as a string. The order of preference is:
// - Values
// - Value
// - ValueFrom.StaticRef
// - ValueFrom.EnvRef
// @return the first value
func (v *Valuable) First() string {
	if len(v.Get()) == 0 {
		return ""
	}

	return v.Get()[0]
}

// String returns the string representation of the Valuable object
// following the order listed on the First() function
func (v Valuable) String() string {
	return v.First()
}

// Contains returns true if the Valuable contains the given value
// @param value is the value to check
// @return true if the Valuable contains the given value
func (v *Valuable) Contains(element string) bool {
	for _, s := range v.Get() {
		if s == element {
			return true
		}
	}
	return false
}

// contains returns true if the Valuable contains the given value.
// This function is private to prevent stack overflow during the initialization
// of the Valuable object.
// @param
// @param value is the value to check
// @return true if the Valuable contains the given value
func contains(slice []string, element string) bool {
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}

// appendCommaListIfAbsent accept a string list separated with commas to append
// to the Values all elements of this list only if element is absent
// of the Values
func appendCommaListIfAbsent(slice []string, commaList string) []string {
	for _, s := range strings.Split(commaList, ",") {
		if contains(slice, s) {
			continue
		}

		slice = append(slice, s)
	}
	return slice
}
