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
	Name             string                `mapstructure:"name"`
	EntrypointURL    string                `mapstructure:"entrypointUrl"`
	Security         []map[string]Security `mapstructure:"security"`
	SecurityPipeline *factory.Pipeline     `mapstructure:"-"`
	Storage          []*StorageSpec        `mapstructure:"storage"`
}

type Security struct {
	ID     string                 `mapstructure:"id"`
	Inputs []*factory.InputConfig `mapstructure:"inputs"`
	Specs  map[string]interface{} `mapstructure:",remain"`
}

type StorageSpec struct {
	Type   string                 `mapstructure:"type"`
	Specs  map[string]interface{} `mapstructure:"specs"`
	Client storage.Pusher
}
