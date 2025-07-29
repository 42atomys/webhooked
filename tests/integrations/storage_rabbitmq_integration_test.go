//go:build integrations

package integration_test

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	StorageTypeRabbitMQ storageType = "rabbitmq"
)

type RabbitMQIntegrationTestSuite struct {
	IntegrationTestSuite
}

func (suite *RabbitMQIntegrationTestSuite) SetupSuite() {
	suite.IntegrationTestSuite.SetupSuite()

	// Initialize RabbitMQ connection
	conn, err := amqp.Dial("amqp://rabbitmq:rabbitmq@rabbitmq:5672/")
	require.NoError(suite.T(), err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	require.NoError(suite.T(), err, "Failed to open channel")

	// Declare test exchange
	err = ch.ExchangeDeclare(
		"webhooks", // name
		"topic",    // type
		false,      // durable
		true,       // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	require.NoError(suite.T(), err, "Failed to declare exchange")

	suite.storages[StorageTypeRabbitMQ] = ch
}

func (suite *RabbitMQIntegrationTestSuite) TearDownSuite() {
	if ch, ok := suite.storages[StorageTypeRabbitMQ].(*amqp.Channel); ok {
		// Delete test queues and exchange
		_, _ = ch.QueueDelete("rabbitmq-basic-storage", true, false, false)
		_, _ = ch.QueueDelete("rabbitmq-formatted-storage", true, false, false)

		_ = ch.ExchangeDelete("webhooks", false, false)
		_ = ch.Close()
	}
	suite.IntegrationTestSuite.TearDownSuite()
}

func (suite *RabbitMQIntegrationTestSuite) TestRabbitMQStorageScenarios() {
	tests := []testInput{
		{
			name:     "rabbitmq-basic-storage",
			endpoint: "/integration/rabbitmq-basic",
			headers: map[string]string{
				"X-Token": "integration-test",
			},
			payload: map[string]any{
				"event_type": "order.placed",
				"order_id":   "ORD-12345",
				"amount":     250.50,
				"timestamp":  "2023-06-28T20:30:00Z",
			},
			expectedResponse: expectedResponse{
				statusCode: 204,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRabbitMQ,
				key:         "integration.rabbitmq-basic",
				data:        `{"event_type":"order.placed","order_id":"ORD-12345","amount":250.5,"timestamp":"2023-06-28T20:30:00Z"}`,
				isJson:      true,
			},
		},
		{
			name:     "rabbitmq-formatted-storage",
			endpoint: "/integration/rabbitmq-formatted",
			headers: map[string]string{
				"X-Token":      "integration-test",
				"X-Delivery":   "rmq456",
				"Content-Type": "application/json",
			},
			payload: map[string]any{
				"action": "notification",
				"type":   "email",
				"to":     "user@example.com",
			},
			expectedResponse: expectedResponse{
				statusCode: 200,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				body: `{"status":"stored","delivery_id":"rmq456"}`,
			},
			expectedStorage: expectedStorage{
				storageType: StorageTypeRabbitMQ,
				key:         "integration.rabbitmq-formatted",
				data:        `rabbitmq|email|user@example.com`,
				isJson:      false,
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.runRabbitMQTest(test)
		})
	}
}

func (suite *RabbitMQIntegrationTestSuite) runRabbitMQTest(test testInput) {
	// Setup consumer before making request
	if test.expectedStorage.storageType == StorageTypeRabbitMQ {
		ch := suite.storages[StorageTypeRabbitMQ].(*amqp.Channel)

		// Create a temporary queue to consume the message
		_, err := ch.QueueDeclare(
			test.name, // name
			false,     // durable
			true,      // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		require.NoError(suite.T(), err, "Failed to declare queue")

		// Bind queue to exchange
		err = ch.QueueBind(
			test.name,                // queue name
			test.expectedStorage.key, // routing key
			"webhooks",               // exchange
			false,
			nil,
		)
		require.NoError(suite.T(), err, "Failed to bind queue")
	}

	suite.doRequest(test)

	// Additional RabbitMQ-specific verification
	if test.expectedStorage.storageType == StorageTypeRabbitMQ {
		ch := suite.storages[StorageTypeRabbitMQ].(*amqp.Channel)

		time.Sleep(100 * time.Millisecond) // Allow time for async storage

		// Try to consume the message
		msgs, err := ch.Consume(
			test.name, // queue
			"",        // consumer
			true,      // auto-ack
			false,     // exclusive
			false,     // no-local
			false,     // no-wait
			nil,       // args
		)
		suite.NoError(err, "Failed to consume messages")

		// Read one message with timeout
		select {
		case msg := <-msgs:
			if test.expectedStorage.isJson {
				suite.JSONEq(test.expectedStorage.data, string(msg.Body), "Data mismatch in RabbitMQ storage")
			} else {
				suite.Equal(test.expectedStorage.data, string(msg.Body), "Data mismatch in RabbitMQ storage")
			}
		case <-time.After(10 * time.Second):
			suite.Fail("Timeout waiting for RabbitMQ message")
		}

		// Clean up queue
		_, err = ch.QueueDelete(test.expectedStorage.key, false, false, false)
		suite.NoError(err, "Failed to delete test queue")
	}
}

func TestRabbitMQIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping RabbitMQ integration tests in short mode")
	}
	suite.Run(t, new(RabbitMQIntegrationTestSuite))
}
