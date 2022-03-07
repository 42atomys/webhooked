package config

import (
	"42stellar.org/webhooks/pkg/factory"
	"42stellar.org/webhooks/pkg/storage"
)

type Configuration struct {
	APIVersion string         `mapstructure:"apiVersion"`
	Specs      []*WebhookSpec `mapstructure:"specs"`
}

type WebhookSpec struct {
	Name              string                              `mapstructure:"name"`
	EntrypointURL     string                              `mapstructure:"entrypointUrl"`
	Security          []map[string]map[string]interface{} `mapstructure:"security"`
	SecurityFactories []*factory.Factory                  `mapstructure:"-"`
	Storage           []*StorageSpec                      `mapstructure:"storage"`
}

type StorageSpec struct {
	Type   string                 `mapstructure:"type"`
	Specs  map[string]interface{} `mapstructure:"specs"`
	Client storage.Pusher
}
