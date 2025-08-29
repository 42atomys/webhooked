package security

import (
	"fmt"
	"reflect"

	"github.com/42atomys/webhooked/internal/hooks"
	"github.com/42atomys/webhooked/security/custom"
	"github.com/42atomys/webhooked/security/github"
	"github.com/42atomys/webhooked/security/noop"
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

	// Map storage type to spec struct
	spec, err := createSpec(securityType)
	if err != nil {
		return nil, fmt.Errorf("error creating spec: %w", err)
	}

	// Decode the specs into the spec struct
	if err := hooks.DecodeField(m, "specs", spec); err != nil {
		return nil, fmt.Errorf("error decoding specs: %w", err)
	}

	// Return the Security struct with the correct spec
	return Security{
		Type:  securityType,
		Specs: spec,
	}, nil
}

// Helper to map storage type to spec struct
func createSpec(securityType string) (Specs, error) {
	switch securityType {
	case "noop":
		return &noop.NoopSecuritySpec{}, nil
	case "github":
		return &github.GitHubSecuritySpec{}, nil
	case "custom":
		return &custom.CustomSecuritySpec{}, nil
	default:
		return nil, fmt.Errorf("unknown security type: %s", securityType)
	}
}
