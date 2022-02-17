package core

import (
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
)

/**
 * Validate the configuration file and her content
 */
func ValidateConfiguration() error {
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

// GetConfig returns the aftual configuration
func GetConfig() *Configuration {
	return config
}
