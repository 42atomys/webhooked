package config

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"42stellar.org/webhooks/pkg/factory"
	"42stellar.org/webhooks/pkg/storage"
)

var (
	currentConfig = &Configuration{}
	// ErrSpecNotFound is returned when the spec is not found
	ErrSpecNotFound = errors.New("spec not found")
)

func Load() error {
	err := viper.Unmarshal(&currentConfig, viper.DecodeHook(factory.DecodeHook))
	if err != nil {
		return err
	}

	for _, spec := range currentConfig.Specs {
		if err := loadSecurityFactory(spec); err != nil {
			return err
		}

		if err = loadStorage(spec); err != nil {
			return fmt.Errorf("storage %s is not valid: %s", spec.Name, err.Error())
		}
	}

	return Validate(currentConfig)
}

// loadSecurityFactory loads the security factory for the given spec
// if an error is occured, return an error
func loadSecurityFactory(spec *WebhookSpec) error {
	spec.SecurityPipeline = factory.NewPipeline()
	for _, security := range spec.Security {
		for securityName, securityConfig := range security {
			f, ok := factory.GetFactoryByName(securityName)
			if !ok {
				return fmt.Errorf("security factory v2 name %s is not valid in %s spec", securityName, spec.Name)
			}

			for _, input := range securityConfig.Inputs {
				f.WithInput(input.Name, input)
			}

			spec.SecurityPipeline.AddFactory(f.WithID(securityConfig.ID).WithConfig(securityConfig.Specs))
		}
	}
	log.Debug().Msgf("%d security factories loaded for spec %s", spec.SecurityPipeline.FactoryCount(), spec.Name)
	return nil
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

	log.Info().Msgf("Load %d configurations", len(config.Specs))
	return nil
}

// loadStorage registers the storage and validate it
// if the storage is not found or an error is occured during the
// initialization or connection, the error is returned during the
// validation
func loadStorage(spec *WebhookSpec) (err error) {
	for _, s := range spec.Storage {
		s.Client, err = storage.Load(s.Type, s.Specs)
		if err != nil {
			return
		}
	}

	log.Debug().Msgf("%d storages loaded for spec %s", len(spec.Storage), spec.Name)
	return
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
			return spec, nil
		}
	}

	log.Warn().Msgf("No spec found for %s endpoint", endpoint)
	return nil, ErrSpecNotFound
}
