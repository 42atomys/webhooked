package format

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
)

func DecodeHook(from reflect.Type, to reflect.Type, data any) (any, error) {
	// Check if we're decoding to a pointer to Formatting
	if from.Kind() != reflect.Map || to != reflect.TypeOf(&Formatting{}) {
		return data, nil
	}

	log.Debug().Msgf("format.DecodeHook: %v -> %v", from, to)
	m, ok := data.(map[string]any)
	if !ok {
		return data, fmt.Errorf("expected map[string]any for Formatting")
	}

	templateStringStr, _ := m["templateString"].(string)
	templatePathStr, _ := m["templatePath"].(string)

	// If both are empty, return nil to avoid unnecessary initialization
	if templateStringStr == "" && templatePathStr == "" {
		return (*Formatting)(nil), nil
	}

	f, err := New(Specs{
		TemplateString: templateStringStr,
		TemplatePath:   templatePathStr,
	})
	if err != nil {
		return nil, err
	}

	return f, nil
}
