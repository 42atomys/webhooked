package rabbitmq

import (
	"bytes"
	"encoding/gob"

	"42stellar.org/webhooks/internal/valuable"
	"github.com/streadway/amqp"
)

// storage is the struct contains client and config
// Run is made from external caller at begins programs
type storage struct {
	config     *config
	client     *amqp.Connection
	channel    *amqp.Channel
	routingKey amqp.Queue
}

// config is the struct contains config for connect client
// Run is made from internal caller
type config struct {
	DatabaseURL      valuable.Valuable
	QueueName        string
	Durable          bool
	DeleteWhenUnused bool
	Exclusive        bool
	NoWait           bool
	Mandatory        bool
	Immediate        bool
	Exchange         string
}

// NewStorage is the function for create new RabbitMQ client storage
// Run is made from external caller at begins programs
// @param config contains config define in the webhooks yaml file
// @return RabbitMQStorage the struct contains client connected and config
// @error an error if the the client is not initialized succesfully
func NewStorage(configRaw map[string]interface{}) (*storage, error) {
	var err error

	newClient := storage{
		config: &config{},
	}

	if err := valuable.Decode(configRaw, &newClient.config); err != nil {
		return nil, err
	}

	if newClient.client, err = amqp.Dial(newClient.config.DatabaseURL.First()); err != nil {
		return nil, err
	}

	if newClient.channel, err = newClient.client.Channel(); err != nil {
		return nil, err
	}

	if newClient.routingKey, err = newClient.channel.QueueDeclare(
		newClient.config.QueueName,
		newClient.config.Durable,
		newClient.config.DeleteWhenUnused,
		newClient.config.Exclusive,
		newClient.config.NoWait,
		nil,
	); err != nil {
		return nil, err
	}

	return &newClient, nil
}

// Name is the function for identified if the storage config is define in the webhooks
// Run is made from external caller
func (c storage) Name() string {
	return "rabbitmq"
}

// Push is the function for push data in the storage
// A run is made from external caller
// @param value that will be pushed
// @return an error if the push failed
func (c storage) Push(value interface{}) error {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return err
	}

	if err := c.channel.Publish(
		c.config.Exchange,
		c.routingKey.Name,
		c.config.Mandatory,
		c.config.Immediate,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(buf.Bytes()),
		}); err != nil {
		return err
	}

	return nil
}
