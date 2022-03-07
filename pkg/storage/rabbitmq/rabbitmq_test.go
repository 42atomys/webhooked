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
		"databaseURL": []int{1},
	})
	assert.Error(suite.T(), err)

	_, err = NewStorage(map[string]interface{}{
		"databaseURL":      "amqp://user:password@127.0.0.1:5672",
		"queueName":        "hello",
		"durable":          false,
		"deleteWhenUnused": false,
		"exclusive":        false,
		"noWait":           false,
		"mandatory":        false,
		"immediate":        false,
	})
	assert.NoError(suite.T(), err)
}

func (suite *RabbitMQSetupTestSuite) TestRabbitMQPush() {
	newClient, _ := NewStorage(map[string]interface{}{
		"databaseURL":      "amqp://user:password@127.0.0.1:5672",
		"queueName":        "hello",
		"durable":          false,
		"deleteWhenUnused": false,
		"exclusive":        false,
		"noWait":           false,
		"mandatory":        false,
		"immediate":        false,
	})
	err := newClient.Push(func() {})
	assert.Error(suite.T(), err)

	newClient, err = NewStorage(map[string]interface{}{
		"databaseURL":      "amqp://user:password@127.0.0.1:5672",
		"queueName":        "hello",
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
