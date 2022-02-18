package config

type Configuration struct {
	APIVersion string         `mapstructure:"apiVersion"`
	Specs      []*WebhookSpec `mapstructure:"specs"`
}

type WebhookSpec struct {
	Name          string                  `mapstructure:"name"`
	EntrypointURL string                  `mapstructure:"entrypointUrl"`
	Security      map[string]SecuritySpec `mapstructure:"security"`
	Storage       map[string]StorageSpec  `mapstructure:"storage"`
}

type SecuritySpec struct{}
type StorageSpec struct{}
