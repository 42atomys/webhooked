package factory

import (
	"fmt"
	"reflect"

	"42stellar.org/webhooks/internal/valuable"
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
	var name = ""
	for k, v := range data.(map[interface{}]interface{}) {
		if fmt.Sprintf("%v", k) == "name" {
			name = fmt.Sprintf("%s", v)
			break
		}
	}

	return &InputConfig{
		Valuable: *v,
		Name:     name,
	}, err
}
