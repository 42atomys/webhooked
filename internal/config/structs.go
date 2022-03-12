package config

import (
	"atomys.codes/webhooked/pkg/factory"
	"atomys.codes/webhooked/pkg/storage"
)

type Configuration struct {
	APIVersion    string         `mapstructure:"apiVersion"`
	Observability Observability  `mapstructure:"observability"`
	Specs         []*WebhookSpec `mapstructure:"specs"`
}

type Observability struct {
	MetricsEnabled bool `mapstructure:"metricsEnabled"`
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
