package config

import (
	"atomys.codes/webhooked/pkg/factory"
	"atomys.codes/webhooked/pkg/storage"
)

// Configuration is the struct contains all the configuration
// defined in the webhooks yaml file
type Configuration struct {
	// APIVerion is the version of the API that will be used
	APIVersion string `mapstructure:"apiVersion"`
	// Observability is the configuration for observability
	Observability Observability `mapstructure:"observability"`
	// Specs is the configuration for the webhooks specs
	Specs []*WebhookSpec `mapstructure:"specs"`
}

// Observability is the struct contains the configuration for observability
// defined in the webhooks yaml file.
type Observability struct {
	// MetricsEnabled is the flag to enable or disable the prometheus metrics
	// endpoint and expose the metrics
	MetricsEnabled bool `mapstructure:"metricsEnabled"`
}

// WebhookSpec is the struct contains the configuration for a webhook spec
// defined in the webhooks yaml file.
type WebhookSpec struct {
	// Name is the name of the webhook spec. It must be unique in the configuration
	// file. It is used to identify the webhook spec in the configuration file
	// and is defined by the user
	Name string `mapstructure:"name"`
	// EntrypointURL is the URL of the entrypoint of the webhook spec. It must
	// be unique in the configuration file. It is defined by the user
	// It is used to identify the webhook spec when receiving a request
	EntrypointURL string `mapstructure:"entrypointUrl"`
	// Security is the configuration for the security of the webhook spec
	// It is defined by the user and can be empty. See HasSecurity() method
	// to know if the webhook spec has security
	Security []map[string]Security `mapstructure:"security"`
	// SecurityPipeline is the security pipeline of the webhook spec
	// It is defined by the configuration loader. This field is not defined
	// by the user and cannot be overridden
	SecurityPipeline *factory.Pipeline `mapstructure:"-"`
	// Storage is the configuration for the storage of the webhook spec
	// It is defined by the user and can be empty.
	Storage []*StorageSpec `mapstructure:"storage"`
}

// Security is the struct contains the configuration for a security
// defined in the webhooks yaml file.
type Security struct {
	// ID is the ID of the security. It must be unique in the configuration
	// file. It is defined by the user and is used to identify the security
	// factory as .Outputs
	ID string `mapstructure:"id"`
	// Inputs is the configuration for the inputs of the security. It is
	// defined by the user and following the specification of the security
	// factory
	Inputs []*factory.InputConfig `mapstructure:"inputs"`
	// Specs is the configuration for the specs of the security. It is
	// defined by the user and following the specification of the security
	// factory
	Specs map[string]interface{} `mapstructure:",remain"`
}

// StorageSpec is the struct contains the configuration for a storage
// defined in the webhooks yaml file.
type StorageSpec struct {
	// Type is the type of the storage. It must be a valid storage type
	// defined in the storage package.
	Type string `mapstructure:"type"`
	// Specs is the configuration for the storage. It is defined by the user
	// following the storage type specification
	Specs map[string]interface{} `mapstructure:"specs"`
	// Client is the storage client. It is defined by the configuration loader
	// and cannot be overridden
	Client storage.Pusher `mapstructure:"-"`
}
