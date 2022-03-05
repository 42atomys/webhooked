package valuable

import (
	"errors"
	"fmt"
	"os"
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

type ValueFromSource struct {
	// StaticRef represents the `staticRef` field of a configuration entry
	// that contains a static value. Can contain a comma separated list
	StaticRef *string `json:"staticRef,omitempty"`
	// EnvRef represents the `envRef` field of a configuration entry
	// that contains a reference to an environment variable
	EnvRef *string `json:"envRef,omitempty"`
}

// SerializeValuable serialize anything to a Valuable
// @param data is the data to serialize
// @return the serialized Valuable
func SerializeValuable(data interface{}) (*Valuable, error) {
	switch t := data.(type) {
	case string:
		return &Valuable{Value: &t}, nil
	case int, float32, float64, bool:
		str := fmt.Sprint(t)
		return &Valuable{Value: &str}, nil
	case nil:
		return &Valuable{}, nil
	default:
		valuable := Valuable{}
		if err := mapstructure.Decode(data, &valuable); err != nil {
			return nil, errors.New("unimplemented valuable type")
		}
		return &valuable, nil
	}
}

// Get returns all values of the Valuable as a slice
// @return the slice of values
func (v *Valuable) Get() []string {
	if v.Values == nil {
		v.Values = make([]string, 0)
	}

	if v.Value != nil && !v.Contains(*v.Value) {
		v.Values = append(v.Values, *v.Value)
	}

	if v.ValueFrom == nil {
		return v.Values
	}

	if v.ValueFrom.StaticRef != nil && !v.Contains(*v.ValueFrom.StaticRef) {
		v.appendCommaListIfAbsent(*v.ValueFrom.StaticRef)
	}

	if v.ValueFrom.EnvRef != nil {
		v.appendCommaListIfAbsent(os.Getenv(*v.ValueFrom.EnvRef))
	}

	return v.Values
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

// Contains returns true if the Valuable contains the given value
// @param value is the value to check
// @return true if the Valuable contains the given value
func (v *Valuable) Contains(element string) bool {
	for _, s := range v.Values {
		if s == element {
			return true
		}
	}
	return false
}

// appendCommaListIfAbsent accept a string list separated with commas to append
// to the Values all elements of this list only if element is absent
// of the Values
func (v *Valuable) appendCommaListIfAbsent(commaList string) {
	for _, s := range strings.Split(commaList, ",") {
		if v.Contains(s) {
			continue
		}

		v.Values = append(v.Values, s)
	}
}
