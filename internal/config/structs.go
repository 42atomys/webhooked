package config

type Configuration struct {
	APIVersion string                 `mapstructure:"apiVersion"`
	Specs      map[string]WebhookSpec `mapstructure:"specs"`
}

type WebhookSpec struct {
	EntrypointURL string                  `mapstructure:"entrypointUrl"`
	Security      map[string]SecuritySpec `mapstructure:"security"`
	Storage       map[string]StorageSpec  `mapstructure:"storage"`
}

type SecuritySpec struct{}
type StorageSpec struct{}
