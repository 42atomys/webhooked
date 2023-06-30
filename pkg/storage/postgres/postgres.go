package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"atomys.codes/webhooked/internal/valuable"
)

// storage is the struct contains client and config
// Run is made from external caller at begins programs
type storage struct {
	client *sql.DB
	config *config
}

// config is the struct contains config for connect client
// Run is made from internal caller
type config struct {
	DatabaseURL valuable.Valuable `mapstructure:"databaseUrl" json:"databaseUrl"`
	// ! Deprecation notice: End of life in v1.0.0
	TableName string `mapstructure:"tableName" json:"tableName"`
	// ! Deprecation notice: End of life in v1.0.0
	DataField string `mapstructure:"dataField" json:"dataField"`

	UseFormattingToPerformQuery bool `mapstructure:"useFormattingToPerformQuery" json:"useFormattingToPerformQuery"`
}

// NewStorage is the function for create new Postgres client storage
// Run is made from external caller at begins programs
// @param config contains config define in the webhooks yaml file
// @return PostgresStorage the struct contains client connected and config
// @return an error if the the client is not initialized successfully
func NewStorage(configRaw map[string]interface{}) (*storage, error) {
	var err error

	newClient := storage{
		config: &config{},
	}

	if err := valuable.Decode(configRaw, &newClient.config); err != nil {
		return nil, err
	}

	// ! Deprecation notice: End of life in v1.0.0
	if newClient.config.TableName != "" || newClient.config.DataField != "" {
		log.Warn().Msg("[DEPRECATION NOTICE] The TableName and DataField are deprecated, please use the formatting feature instead")
	}

	if newClient.client, err = sql.Open("postgres", newClient.config.DatabaseURL.First()); err != nil {
		return nil, err
	}

	return &newClient, nil
}

// Name is the function for identified if the storage config is define in the webhooks
// Run is made from external caller
func (c storage) Name() string {
	return "postgres"
}

// Push is the function for push data in the storage.
// The data is formatted with the formatting feature and be serialized by the
// client with "toSql" method
// A run is made from external caller
// @param value that will be pushed
// @return an error if the push failed
func (c storage) Push(value interface{}) error {
	// ! Deprecation notice: End of life in v1.0.0
	if !c.config.UseFormattingToPerformQuery {
		request := fmt.Sprintf("INSERT INTO %s(%s) VALUES ($1)", c.config.TableName, c.config.DataField)
		if _, err := c.client.Query(request, value); err != nil {
			return err
		}
		return nil
	}

	if _, err := c.client.Query(value.(string)); err != nil {
		return err
	}

	return nil
}
