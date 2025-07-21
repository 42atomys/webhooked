// Package hooks provides a list of helpers functions to manipulates hooks
package hooks

import (
	"fmt"

	"github.com/42atomys/webhooked/format"
	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/go-viper/mapstructure/v2"
)

// DecodeField will decode a field from a map by looking on posible valuable or
// formating hooks
func DecodeField(data map[string]any, key string, result any) error {
	if _, exists := data[key]; !exists {
		return nil
	}

	fieldData, ok := data[key].(map[string]any)
	if !ok {
		return fmt.Errorf("%s must be a map", key)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			valuable.MapToValuableHookFunc(),
			format.DecodeHook,
		),
		Result:  result,
		TagName: "json",
	})
	if err != nil {
		return fmt.Errorf("error creating decoder: %w", err)
	}

	return decoder.Decode(fieldData)
}
