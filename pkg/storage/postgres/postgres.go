package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"atomys.codes/webhooked/internal/valuable"
	"atomys.codes/webhooked/pkg/formatting"
)

// storage is the struct contains client and config
// Run is made from external caller at begins programs
type storage struct {
	client *sqlx.DB
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
	// The query to perform on the database with named arguments
	Query string `mapstructure:"query" json:"query"`
	// The arguments to use in the query with the formatting feature (see pkg/formatting)
	Args map[string]string `mapstructure:"args" json:"args"`
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

	if newClient.config.UseFormattingToPerformQuery {
		if newClient.config.TableName != "" || newClient.config.DataField != "" {
			return nil, fmt.Errorf("the formatting feature is enabled, the TableName and DataField are deprecated and cannot be used in the same time")
		}

		if newClient.config.Query == "" {
			return nil, fmt.Errorf("the query is required when the formatting feature is enabled")
		}

		if newClient.config.Args == nil {
			newClient.config.Args = make(map[string]string, 0)
		}
	}

	if newClient.client, err = sqlx.Open("postgres", newClient.config.DatabaseURL.First()); err != nil {
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
func (c storage) Push(ctx context.Context, value []byte) error {
	// ! Deprecation notice: End of life in v1.0.0
	if !c.config.UseFormattingToPerformQuery {
		request := fmt.Sprintf("INSERT INTO %s(%s) VALUES ($1)", c.config.TableName, c.config.DataField)
		if _, err := c.client.Query(request, value); err != nil {
			return err
		}
		return nil
	}

	formatter, err := formatting.FromContext(ctx)
	if err != nil {
		return err
	}

	stmt, err := c.client.PrepareNamedContext(ctx, c.config.Query)
	if err != nil {
		return err
	}

	var namedArgs = make(map[string]interface{}, 0)
	for name, template := range c.config.Args {
		value, err := formatter.
			WithPayload(value).
			WithTemplate(template).
			WithData("FieldName", name).
			Render()
		if err != nil {
			return err
		}

		namedArgs[name] = value
	}

	_, err = stmt.QueryContext(ctx, namedArgs)
	return err
}
