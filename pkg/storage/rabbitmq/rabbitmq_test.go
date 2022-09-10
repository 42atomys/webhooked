package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RabbitMQSetupTestSuite struct {
	suite.Suite
}

func (suite *RabbitMQSetupTestSuite) TestRabbitMQName() {
	newRabbitMQ := storage{}
	assert.Equal(suite.T(), "rabbitmq", newRabbitMQ.Name())
}

func (suite *RabbitMQSetupTestSuite) TestRabbitMQNewStorage() {
	_, err := NewStorage(map[string]interface{}{
		"databaseUrl": []int{1},
	})
	assert.Error(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"databaseUrl":      "amqp://user:password@127.0.0.1:5672",
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
		"databaseUrl":      "amqp://user:password@127.0.0.1:5672",
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

	err = newClient.Push("Hello")
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

func TestReconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("rabbitmq testing is skiped in short version of test")
		return
	}

	newClient, err := NewStorage(map[string]interface{}{
		"databaseUrl":      "amqp://user:password@127.0.0.1:5672",
		"queueName":        "hello",
		"contentType":      "text/plain",
		"durable":          false,
		"deleteWhenUnused": false,
		"exclusive":        false,
		"noWait":           false,
		"mandatory":        false,
		"immediate":        false,
	})
	assert.NoError(t, err)

	assert.NoError(t, newClient.Push("Hello"))
	assert.NoError(t, newClient.client.Close())
	assert.NoError(t, newClient.Push("Hello"))
	assert.NoError(t, newClient.channel.Close())
	assert.NoError(t, newClient.Push("Hello"))
}
