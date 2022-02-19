package config

import "42stellar.org/webhooks/pkg/factory"

type Configuration struct {
	APIVersion string         `mapstructure:"apiVersion"`
	Specs      []*WebhookSpec `mapstructure:"specs"`
}

type WebhookSpec struct {
	Name              string                              `mapstructure:"name"`
	EntrypointURL     string                              `mapstructure:"entrypointUrl"`
	Security          []map[string]map[string]interface{} `mapstructure:"security"`
	SecurityFactories []*factory.Factory                  `mapstructure:"-"`
	Storage           map[string]StorageSpec              `mapstructure:"storage"`
}

type StorageSpec struct{}
