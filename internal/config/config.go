package config

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Configuration struct {
	APIVersion string                       `mapstructure:"apiVersion"`
	Specs      map[string]ConfigurationSpec `mapstructure:"specs"`
}

type ConfigurationSpec struct {
	EntrypointURL string                  `mapstructure:"entrypointUrl"`
	Security      map[string]SecuritySpec `mapstructure:"security"`
	Storage       map[string]StorageSpec  `mapstructure:"storage"`
}

type SecuritySpec struct{}
type StorageSpec struct{}

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

// GetEntry returns the spec for the given name, if no entry
// is found, ErrSpecNotFound is returned
func (c *Configuration) GetEntry(name string) (*ConfigurationSpec, error) {
	spec, ok := c.Specs[name]
	if !ok {
		log.Error().Err(ErrSpecNotFound).Msgf("Spec %s not found", name)
		return nil, ErrSpecNotFound
	}

	return &spec, nil
}

// GetEntryByEndpoint returns the spec for the given endpoint, if no entry
// is found, ErrSpecNotFound is returned
func (c *Configuration) GetEntryByEndpoint(endpoint string) (*ConfigurationSpec, error) {
	for _, spec := range c.Specs {
		if spec.EntrypointURL == endpoint {
			log.Warn().Msgf("No spec found for %s endpoint", endpoint)
			return &spec, nil
		}
	}

	return nil, ErrSpecNotFound
}
