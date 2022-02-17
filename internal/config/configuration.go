package config

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	config = &Configuration{}
	// ErrSpecNotFound is returned when the spec is not found
	ErrSpecNotFound = errors.New("spec not found")
)

/**
 * Validate the configuration file and her content
 */
func Validate() error {
	err := viper.Unmarshal(&config)
	if err != nil {
		return err
	}

	// TODO Validation of the configuration if necessary
	// for name, spec := range config.Specs {
	// 	log.Debug().Str("name", name).Msgf("Load spec: %+v", spec)
	// }

	log.Debug().Msgf("Load %d configurations", len(config.Specs))
	return nil
}

// Current returns the aftual configuration
func Current() *Configuration {
	return config
}

// GetSpec returns the spec for the given name, if no entry
// is found, ErrSpecNotFound is returned
func (c *Configuration) GetSpec(name string) (*WebhookSpec, error) {
	spec, ok := c.Specs[name]
	if !ok {
		log.Error().Err(ErrSpecNotFound).Msgf("Spec %s not found", name)
		return nil, ErrSpecNotFound
	}

	return &spec, nil
}

// GetSpecByEndpoint returns the spec for the given endpoint, if no entry
// is found, ErrSpecNotFound is returned
func (c *Configuration) GetSpecByEndpoint(endpoint string) (*WebhookSpec, error) {
	for _, spec := range c.Specs {
		if spec.EntrypointURL == endpoint {
			log.Warn().Msgf("No spec found for %s endpoint", endpoint)
			return &spec, nil
		}
	}

	return nil, ErrSpecNotFound
}
