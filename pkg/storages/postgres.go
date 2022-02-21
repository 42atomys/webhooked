package storages

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
)

type PostgresStorage struct {
	client *sql.DB
	config *postgresConfig
}

type postgresConfig struct {
	DatabaseURL string
	TableName   string
	DataField   string
}

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

func (c PostgresStorage) Name() string {
	return "postgres"
}

func (c PostgresStorage) Push(value interface{}) error {
	request := fmt.Sprintf("INSERT INTO %s(%s) VALUES ('%s')", c.config.TableName, c.config.DataField, value)
	if _, err := c.client.Query(request); err != nil {
		return err
	}

	return nil
}
