package config

import "42stellar.org/webhooks/pkg/core"

type Configuration struct {
	APIVersion string         `mapstructure:"apiVersion"`
	Specs      []*WebhookSpec `mapstructure:"specs"`
}

type WebhookSpec struct {
	Name          string                  `mapstructure:"name"`
	EntrypointURL string                  `mapstructure:"entrypointUrl"`
	Security      map[string]SecuritySpec `mapstructure:"security"`
	Storages      []*StorageSpec          `mapstructure:"storages"`
}

type SecuritySpec struct{}

type StorageSpec struct {
	Type   string                 `mapstructure:"type"`
	Specs  map[string]interface{} `mapstructure:"specs"`
	Client core.IStorage
}
