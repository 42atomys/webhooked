package storage

import (
	"fmt"
	"reflect"

	"github.com/42atomys/webhooked/format"
	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/42atomys/webhooked/storage/noop"
	"github.com/42atomys/webhooked/storage/postgres"
	"github.com/42atomys/webhooked/storage/rabbitmq"
	"github.com/42atomys/webhooked/storage/redis"
	"github.com/go-viper/mapstructure/v2"
)

func DecodeHook(from reflect.Type, to reflect.Type, data any) (any, error) {
	if from.Kind() != reflect.Map || to != reflect.TypeOf(Storage{}) {
		return data, nil
	}

	m, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any for Storage")
	}

	// Extract and validate the type
	storageType, ok := m["type"].(string)
	if !ok {
		return nil, fmt.Errorf("storage type must be a string")
	}

	// Map storage type to spec struct
	spec, err := createSpec(storageType)
	if err != nil {
		return nil, err
	}

	// Decode specs
	if err := decodeField(m, "specs", spec); err != nil {
		return nil, fmt.Errorf("error decoding specs: %w", err)
	}

	// Decode formatting
	// formatting := format.Formatting{}
	// if err := decodeField(m, "formatting", &formatting); err != nil {
	// 	return nil, fmt.Errorf("error decoding formatting: %w", err)
	// }

	return Storage{
		Type: storageType,
		// Formatting: formatting,
		Specs: spec,
	}, nil
}

// Helper to map storage type to spec struct
func createSpec(storageType string) (Specs, error) {
	switch storageType {
	case "noop":
		return &noop.NoopStorageSpec{}, nil
	case "postgres":
		return &postgres.PostgresStorageSpec{}, nil
	case "redis":
		return &redis.RedisStorageSpec{}, nil
	case "rabbitmq":
		return &rabbitmq.RabbitmqStorageSpec{}, nil
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}

// Helper to decode a field from the map
func decodeField(data map[string]any, key string, result any) error {
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
		return err
	}

	return decoder.Decode(fieldData)
}
