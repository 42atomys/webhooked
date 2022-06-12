package factory

import (
	"fmt"
	"reflect"

	"atomys.codes/webhooked/internal/valuable"
)

// DecodeHook is a mapstructure.DecodeHook that serializes
// the given data into a InputConfig with a name and a Valuable object.
// mapstructure cannot nested objects, so we need to serialize the
// data into a map[string]interface{} and then deserialize it into
// a InputConfig.
//
// @see https://pkg.go.dev/github.com/mitchellh/mapstructure#DecodeHookFunc for more details.
func DecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(InputConfig{}) {
		return data, nil
	}

	v, err := valuable.SerializeValuable(data)
	if err != nil {
		return nil, err
	}

	var name = ""
	for k, v2 := range rangeOverInterfaceMap(data) {
		if fmt.Sprintf("%v", k) == "name" {
			name = fmt.Sprintf("%s", v2)
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return &InputConfig{
		Valuable: *v,
		Name:     name,
	}, nil
}

// rangeOverInterfaceMap iterates over the given interface map to convert it
// into a map[string]interface{}. This is needed because mapstructure cannot
// handle objects that are not of type map[string]interface{} for obscure reasons.
func rangeOverInterfaceMap(data interface{}) map[string]interface{} {
	transformedData, ok := data.(map[string]interface{})
	if !ok {
		transformedData = make(map[string]interface{})
		for k, v := range data.(map[interface{}]interface{}) {
			transformedData[fmt.Sprintf("%v", k)] = v
		}
	}

	return transformedData
}
