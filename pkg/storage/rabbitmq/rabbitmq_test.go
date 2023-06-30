package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RabbitMQSetupTestSuite struct {
	suite.Suite
	amqpUrl string
}

func (suite *RabbitMQSetupTestSuite) TestRabbitMQName() {
	newRabbitMQ := storage{}
	assert.Equal(suite.T(), "rabbitmq", newRabbitMQ.Name())
}

// Create Table for running test
func (suite *RabbitMQSetupTestSuite) BeforeTest(suiteName, testName string) {
	suite.amqpUrl = fmt.Sprintf(
		"amqp://%s:%s@%s:%s",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)
}

func (suite *RabbitMQSetupTestSuite) TestRabbitMQNewStorage() {
	_, err := NewStorage(map[string]interface{}{
		"databaseUrl": []int{1},
	})
	assert.Error(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"databaseUrl":      suite.amqpUrl,
		"queueName":        "hello",
		"durable":          false,
		"deleteWhenUnused": false,
		"exclusive":        false,
		"noWait":           false,
		"mandatory":        false,
		"immediate":        false,
	})
	assert.NoError(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"databaseUrl": "amqp://user:",
	})
	assert.Error(suite.T(), err)
}

func (suite *RabbitMQSetupTestSuite) TestRabbitMQPush() {
	newClient, err := NewStorage(map[string]interface{}{
		"databaseUrl":      suite.amqpUrl,
		"queueName":        "hello",
		"contentType":      "text/plain",
		"durable":          false,
		"deleteWhenUnused": false,
		"exclusive":        false,
		"noWait":           false,
		"mandatory":        false,
		"immediate":        false,
	})
	assert.NoError(suite.T(), err)

	err = newClient.Push(context.Background(), []byte("Hello"))
	assert.NoError(suite.T(), err)
}

func TestRunRabbitMQPush(t *testing.T) {
	if testing.Short() {
		t.Skip("rabbitmq testing is skiped in short version of test")
		return
	}

	suite.Run(t, new(RabbitMQSetupTestSuite))
}

func TestContentType(t *testing.T) {
	assert.Equal(t, "text/plain", (&config{}).ContentType())
	assert.Equal(t, "text/plain", (&config{DefinedContentType: ""}).ContentType())
	assert.Equal(t, "application/json", (&config{DefinedContentType: "application/json"}).ContentType())
}

func (suite *RabbitMQSetupTestSuite) TestReconnect() {
	if testing.Short() {
		suite.T().Skip("rabbitmq testing is skiped in short version of test")
		return
	}

	newClient, err := NewStorage(map[string]interface{}{
		"databaseUrl":      suite.amqpUrl,
		"queueName":        "hello",
		"contentType":      "text/plain",
		"durable":          false,
		"deleteWhenUnused": false,
		"exclusive":        false,
		"noWait":           false,
		"mandatory":        false,
		"immediate":        false,
	})
	assert.NoError(suite.T(), err)

	assert.NoError(suite.T(), newClient.Push(context.Background(), []byte("Hello")))
	assert.NoError(suite.T(), newClient.client.Close())
	assert.NoError(suite.T(), newClient.Push(context.Background(), []byte("Hello")))
	assert.NoError(suite.T(), newClient.channel.Close())
	assert.NoError(suite.T(), newClient.Push(context.Background(), []byte("Hello")))
}
