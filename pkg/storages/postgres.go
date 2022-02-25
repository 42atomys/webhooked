package storages

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
)

// PostgresStorage is the struct contains client and config
// Run is made from external caller at begins programs
type PostgresStorage struct {
	client *sql.DB
	config *postgresConfig
}

// postgresConfig is the struct contains config for connect client
// Run is made from internal caller
type postgresConfig struct {
	DatabaseURL string
	TableName   string
	DataField   string
}

// NewPostgresStorage is the function for create new Postgres client storage
// Run is made from external caller at begins programs
// @param config contains config define in the webhooks yaml file
// @return PostgresStorage the struct contains client connected and config
// @error an error if the the client is not initialized succesfully
func NewPostgresStorage(config map[string]interface{}) (*PostgresStorage, error) {
	var err error

	newClient := PostgresStorage{
		config: &postgresConfig{},
	}

	if err := mapstructure.Decode(config, &newClient.config); err != nil {
		return nil, err
	}

	if newClient.client, err = sql.Open("postgres", newClient.config.DatabaseURL); err != nil {
		return nil, err
	}

	return &newClient, nil
}

// Name is the function for identified if the storage config is define in the webhooks
// Run is made from external caller
func (c PostgresStorage) Name() string {
	return "postgres"
}

// Push is the function for push data in the storage
// A run is made from external caller
// @param value that will be pushed
// @return an error if the push failed
func (c PostgresStorage) Push(value interface{}) error {
	request := fmt.Sprintf("INSERT INTO %s(%s) VALUES ('%s')", c.config.TableName, c.config.DataField, value)
	if _, err := c.client.Query(request); err != nil {
		return err
	}

	return nil
}
