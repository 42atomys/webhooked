package integration_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	StorageTypePostgres storageType = "postgres"
)

type PostgresIntegrationTestSuite struct {
	IntegrationTestSuite
}

func (suite *PostgresIntegrationTestSuite) SetupSuite() {
	suite.IntegrationTestSuite.SetupSuite()

	// Initialize PostgreSQL client
	dsn := "postgres://postgres:postgres@postgres:5432/webhooked_test?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(suite.T(), err, "Failed to connect to PostgreSQL")

	err = db.Ping()
	require.NoError(suite.T(), err, "Failed to ping PostgreSQL")

	// Create database if it doesn't exist
	_, err = db.Exec(`
	DO $$
	BEGIN
		IF NOT EXISTS (
				SELECT FROM pg_database WHERE datname = 'webhooked_test'
		) THEN
				CREATE DATABASE webhooked_test;
		END IF;
	END
	$$;
	`)
	require.NoError(suite.T(), err, "Failed to create database if it does not exist")

	// Clean up test table
	_, err = db.Exec("DROP TABLE IF EXISTS webhook_events")
	require.NoError(suite.T(), err, "Failed to drop test table")

	// Create test table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS webhook_events (
		id SERIAL PRIMARY KEY,
		webhook_name TEXT NOT NULL,
		payload TEXT NOT NULL,
		received_at TIMESTAMPTZ NOT NULL
	);
	`)
	require.NoError(suite.T(), err, "Failed to create test table")

	suite.storages[StorageTypePostgres] = db
}

func (suite *PostgresIntegrationTestSuite) TearDownSuite() {
	if db, ok := suite.storages[StorageTypePostgres].(*sql.DB); ok {
		db.Close()
	}
	suite.IntegrationTestSuite.TearDownSuite()
}

func (suite *PostgresIntegrationTestSuite) TestPostgresStorageScenarios() {
	tests := []testInput{
		{
			name:     "postgres-basic-storage",
			endpoint: "/integration/postgres-basic",
			headers: map[string]string{
				"X-Token": "integration-test",
			},
			payload: map[string]any{
				"event_type": "user.created",
				"user_id":    12345,
				"timestamp":  "2023-06-28T18:30:00Z",
			},
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypePostgres,
				key:         "postgres-basic",
				data:        `{"event_type":"user.created","user_id":12345,"timestamp":"2023-06-28T18:30:00Z"}`,
				isJson:      true,
			},
		},
		{
			name:     "postgres-formatted-storage",
			endpoint: "/integration/postgres-formatted",
			headers: map[string]string{
				"X-Token":      "integration-test",
				"X-Delivery":   "abc123",
				"Content-Type": "application/json",
			},
			payload: map[string]any{
				"action":   "purchase",
				"amount":   99.99,
				"currency": "USD",
			},
			expectedResponse: expectedResponse{
				statusCode: 200,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: `{"status":"stored","delivery_id":"abc123"}`,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypePostgres,
				key:         "postgres-formatted",
				data:        `{"webhook":"postgres-formatted","delivery_id":"abc123","event":{"action":"purchase","amount":99.99,"currency":"USD"}}`,
				isJson:      true,
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.runPostgresTest(test)
		})
	}
}

func (suite *PostgresIntegrationTestSuite) runPostgresTest(test testInput) {
	suite.doRequest(test)

	// Additional PostgreSQL-specific verification
	if test.expectedStorage.storageType == StorageTypePostgres {
		db := suite.storages[StorageTypePostgres].(*sql.DB)

		time.Sleep(100 * time.Millisecond) // Allow time for async storage

		var payload string
		var webhookName string
		err := db.QueryRow("SELECT webhook_name, payload FROM webhook_events WHERE webhook_name = $1 ORDER BY id DESC LIMIT 1",
			test.expectedStorage.key).Scan(&webhookName, &payload)

		suite.NoError(err, "Failed to query PostgreSQL storage")
		suite.Equal(test.expectedStorage.key, webhookName, "Webhook name mismatch")
		if test.expectedStorage.isJson {
			suite.JSONEq(test.expectedStorage.data, payload, "Data mismatch in PostgreSQL storage")
		} else {
			suite.Equal(test.expectedStorage.data, payload, "Data mismatch in PostgreSQL storage")
		}

		// Clean up for next test
		_, err = db.Exec("DELETE FROM webhook_events WHERE webhook_name = $1", test.expectedStorage.key)
		suite.NoError(err, "Failed to clean up test data")
	}
}

func TestPostgresIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration tests in short mode")
	}
	suite.Run(t, new(PostgresIntegrationTestSuite))
}
