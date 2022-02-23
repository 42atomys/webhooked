package config

import (
	"errors"
	"fmt"

	"42stellar.org/webhooks/pkg/factory"
	"42stellar.org/webhooks/pkg/storages"
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

	for _, spec := range currentConfig.Specs {
		if err := LoadSecurityFactory(spec); err != nil {
			return err
		}
	}

	return Validate(currentConfig)
}

// LoadSecurityFactory loads the security factory for the given spec
// if an error is occured, return an error
func LoadSecurityFactory(spec *WebhookSpec) error {
	for _, security := range spec.Security {
		for securityName, securityConfig := range security {
			factoryFunc, ok := factory.GetFunctionByName(securityName)
			if !ok {
				return fmt.Errorf("security factory name %s is not valid in %s spec", securityName, spec.Name)
			}
			log.Debug().Msgf("security factory name %s is valid in %s spec", securityName, spec.Name)
			spec.SecurityFactories = append(spec.SecurityFactories, &factory.Factory{Name: securityName, Fn: factoryFunc, Config: securityConfig})
		}
	}
	log.Debug().Msgf("%d security factories loaded for spec %s", len(spec.SecurityFactories), spec.Name)
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

		// Validate the storage
		if err := registerAndvalidateStorage(spec); err != nil {
			return fmt.Errorf("storage %s is not valid: %s", spec.Name, err.Error())
		}
	}

	log.Info().Msgf("Load %d configurations", len(config.Specs))
	return nil
}

// registerAndvalidateStorage registers the storage and validate it
// if the storage is not found or an error is occured during the
// initialization or connection, the error is returned during the
// validation
func registerAndvalidateStorage(spec *WebhookSpec) error {
	var err error
	for _, storage := range spec.Storage {
		switch storage.Type {
		case "redis":
			storage.Client, err = storages.NewRedisStorage(storage.Specs)
			if err != nil {
				return err
			}

		case "postgres":
			storage.Client, err = storages.NewPostgresStorage(storage.Specs)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("storage %s is undefined", storage.Type)
		}
	}
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
			return spec, nil
		}
	}

	log.Warn().Msgf("No spec found for %s endpoint", endpoint)
	return nil, ErrSpecNotFound
}
