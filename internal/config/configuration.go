package config

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	currentConfig = &Configuration{}
	// ErrSpecNotFound is returned when the spec is not found
	ErrSpecNotFound = errors.New("spec not found")
)

func Load() error {
	err := viper.Unmarshal(&currentConfig)
	if err != nil {
		return err
	}

	return Validate(currentConfig)
}

/**
 * Validate the configuration file and her content
 */
func Validate(config *Configuration) error {
	var uniquenessName = make(map[string]bool)
	var uniquenessUrl = make(map[string]bool)

	for _, spec := range config.Specs {
		log.Debug().Str("name", spec.Name).Msgf("Load spec: %+v", spec)

		// Validate the uniqueness of all name
		if _, ok := uniquenessName[spec.Name]; ok {
			return fmt.Errorf("name %s is not unique", spec.Name)
		}
		uniquenessName[spec.Name] = true

		// Validate the uniqueness of all entrypoints
		if _, ok := uniquenessUrl[spec.EntrypointURL]; ok {
			return fmt.Errorf("entrypointUrl %s is not unique", spec.EntrypointURL)
		}
		uniquenessUrl[spec.EntrypointURL] = true
	}

	log.Debug().Msgf("Load %d configurations", len(config.Specs))
	return nil
}

// Current returns the aftual configuration
func Current() *Configuration {
	return currentConfig
}

// GetSpec returns the spec for the given name, if no entry
// is found, ErrSpecNotFound is returned
func (c *Configuration) GetSpec(name string) (*WebhookSpec, error) {
	for _, spec := range c.Specs {
		if spec.Name == name {
			return spec, nil
		}
	}

	log.Error().Err(ErrSpecNotFound).Msgf("Spec %s not found", name)
	return nil, ErrSpecNotFound

}

// GetSpecByEndpoint returns the spec for the given endpoint, if no entry
// is found, ErrSpecNotFound is returned
func (c *Configuration) GetSpecByEndpoint(endpoint string) (*WebhookSpec, error) {
	for _, spec := range c.Specs {
		if spec.EntrypointURL == endpoint {
			log.Warn().Msgf("No spec found for %s endpoint", endpoint)
			return spec, nil
		}
	}

	return nil, ErrSpecNotFound
}