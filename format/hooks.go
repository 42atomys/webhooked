package format

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
)

func DecodeHook(from reflect.Type, to reflect.Type, data any) (any, error) {
	if from.Kind() != reflect.Map || to != reflect.TypeOf(Formatting{}) {
		return data, nil
	}

	log.Debug().Msgf("format.DecodeHook: %v -> %v", from, to)
	m, ok := data.(map[string]any)
	if !ok {
		return data, fmt.Errorf("expected map[string]any for Formatting")
	}

	templateStringStr, _ := m["templateString"].(string)
	templatePathStr, _ := m["templatePath"].(string)

	f, err := New(templateStringStr, templatePathStr)
	if err != nil {
		return nil, err
	}

	return *f, nil
}
