package valuable

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// Decode decodes the given data into the given result.
// In case of the target Type if a Valuable, we serialize it with
// `SerializeValuable` func.
// @param input is the data to decode
// @param output is the result of the decoding
// @return an error if the decoding failed
func Decode(input, output interface{}) (err error) {
	var decoder *mapstructure.Decoder

	decoder, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     output,
		DecodeHook: valuableDecodeHook,
	})
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// valuableDecodeHook is a mapstructure.DecodeHook that serializes
// the given data into a Valuable.
func valuableDecodeHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(Valuable{}) {
		return data, nil
	}

	return SerializeValuable(data)
}
