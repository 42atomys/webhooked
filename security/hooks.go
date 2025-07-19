package security

import (
	"fmt"
	"reflect"

	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/42atomys/webhooked/security/custom"
	"github.com/42atomys/webhooked/security/github"
	"github.com/42atomys/webhooked/security/noop"
	"github.com/go-viper/mapstructure/v2"
	"github.com/rs/zerolog/log"
)

func DecodeHook(from reflect.Type, to reflect.Type, data any) (any, error) {
	if from.Kind() != reflect.Map || to != reflect.TypeOf(Security{}) {
		return data, nil
	}

	log.Debug().Msgf("security.DecodeHook: %v -> %v", from, to)
	m, ok := data.(map[string]any)
	if !ok {
		return data, fmt.Errorf("expected map[string]any for Security")
	}

	// Extract the type
	securityType, ok := m["type"].(string)
	if !ok {
		return data, fmt.Errorf("security type must be a string")
	}

	// Depending on the type, create the appropriate spec struct
	var spec Specs
	switch securityType {
	case "noop":
		spec = &noop.NoopSecuritySpec{}
	case "github":
		spec = &github.GitHubSecuritySpec{}
	case "custom":
		spec = &custom.CustomSecuritySpec{}
	default:
		return data, fmt.Errorf("unknown security type: %s", securityType)
	}

	// Decode the specs into the spec struct
	specsData, ok := m["specs"].(map[string]any)
	if !ok {
		return data, fmt.Errorf("specs must be a map, got %v", m["specs"])
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			valuable.MapToValuableHookFunc(),
		),
		Result:  spec,
		TagName: "json",
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(specsData); err != nil {
		return nil, err
	}

	// Return the Security struct with the correct spec
	return Security{
		Type:  securityType,
		Specs: spec,
	}, nil
}
