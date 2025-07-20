package postgres

import (
	"context"
	"fmt"

	"github.com/42atomys/webhooked/format"
	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresStorageSpec struct {
	DatabaseURL valuable.Valuable `mapstructure:"databaseUrl" json:"databaseUrl"`
	Query       string            `mapstructure:"query" json:"query"`
	Args        map[string]string `mapstructure:"args" json:"args"`

	client     *sqlx.DB
	formatters map[string]*format.Formatting // map of formatters keyed by arg name
}

func (s *PostgresStorageSpec) EnsureConfigurationCompleteness() error {
	if s.DatabaseURL.First() == "" {
		return fmt.Errorf("databaseUrl is required")
	}

	if s.Query == "" {
		return fmt.Errorf("query is required")
	}

	if s.Args == nil {
		s.Args = make(map[string]string, 0)
	}

	return nil
}

func (s *PostgresStorageSpec) Initialize() error {
	var err error

	if s.client, err = sqlx.Open("postgres", s.DatabaseURL.First()); err != nil {
		return err
	}

	for name, template := range s.Args {
		formatter, err := format.New(format.Specs{TemplateString: template})
		if err != nil {
			return fmt.Errorf("error initializing formatter for %s: %w", name, err)
		}

		s.formatters[name] = formatter
	}

	return nil
}

func (s *PostgresStorageSpec) Store(ctx context.Context, value []byte) error {
	stmt, err := s.client.PrepareNamedContext(ctx, s.Query)
	if err != nil {
		return err
	}

	var namedArgs = make(map[string]interface{}, 0)
	for name := range s.Args {
		value, err := s.formatters[name].Format(ctx, map[string]any{
			"FieldName": name,
		})
		if err != nil {
			return err
		}
		namedArgs[name] = value
	}

	_, err = stmt.QueryContext(ctx, namedArgs)
	return err

}
