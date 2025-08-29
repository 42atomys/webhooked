package valuable

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/rs/zerolog/log"
)

// Decode decodes the given data into the given result.
// In case of the target Type if a Valuable, we serialize it with
// `SerializeValuable` func.
// @param input is the data to decode
// @param output is the result of the decoding
// @return an error if the decoding failed
func Decode(input, output any) (err error) {
	var decoder *mapstructure.Decoder

	decoder, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     output,
		DecodeHook: MapToValuableHookFunc(),
	})
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func MapToValuableHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(Valuable{}) {
			return data, nil
		}

		log.Debug().Msgf("MapToValuableHookFunc: %v -> %v", f, t)
		return Serialize(data)
	}
}
