package rabbitmq

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"

	"atomys.codes/webhooked/internal/valuable"
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
	DatabaseURL        valuable.Valuable `mapstructure:"databaseUrl" json:"databaseUrl"`
	QueueName          string            `mapstructure:"queueName" json:"queueName"`
	DefinedContentType string            `mapstructure:"contentType" json:"contentType"`
	Durable            bool              `mapstructure:"durable" json:"durable"`
	DeleteWhenUnused   bool              `mapstructure:"deleteWhenUnused" json:"deleteWhenUnused"`
	Exclusive          bool              `mapstructure:"exclusive" json:"exclusive"`
	NoWait             bool              `mapstructure:"noWait" json:"noWait"`
	Mandatory          bool              `mapstructure:"mandatory" json:"mandatory"`
	Immediate          bool              `mapstructure:"immediate" json:"immediate"`
	Exchange           string            `mapstructure:"exchange" json:"exchange"`
}

const maxAttempt = 5

// ContentType is the function for get content type used to push data in the
// storage. When no content type is defined, the default one is used instead
// Default: text/plain
func (c *config) ContentType() string {
	if c.DefinedContentType != "" {
		return c.DefinedContentType
	}

	return "text/plain"
}

// NewStorage is the function for create new RabbitMQ client storage
// Run is made from external caller at begins programs
// @param config contains config define in the webhooks yaml file
// @return RabbitMQStorage the struct contains client connected and config
// @return an error if the the client is not initialized successfully
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

	go func() {
		for {
			reason := <-newClient.client.NotifyClose(make(chan *amqp.Error))
			log.Warn().Msgf("connection to rabbitmq closed, reason: %v", reason)

			newClient.reconnect()
		}
	}()

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
func (c *storage) Name() string {
	return "rabbitmq"
}

// Push is the function for push data in the storage
// A run is made from external caller
// @param value that will be pushed
// @return an error if the push failed
func (c *storage) Push(ctx context.Context, value []byte) error {
	for attempt := 0; attempt < maxAttempt; attempt++ {
		err := c.channel.Publish(
			c.config.Exchange,
			c.routingKey.Name,
			c.config.Mandatory,
			c.config.Immediate,
			amqp.Publishing{
				ContentType: c.config.ContentType(),
				Body:        value,
			})

		if err != nil {
			if errors.Is(err, amqp.ErrClosed) {
				log.Warn().Err(err).Msg("connection to rabbitmq closed. reconnecting...")
				c.reconnect()
				continue
			} else {
				return err
			}
		}
		return nil
	}

	return errors.New("max attempt to publish reached")
}

// reconnect is the function to reconnect to the amqp server if the connection
// is lost. It will try to reconnect every seconds until it succeed to connect
func (c *storage) reconnect() {
	for {
		// wait 1s for reconnect
		time.Sleep(time.Second)

		conn, err := amqp.Dial(c.config.DatabaseURL.First())
		if err == nil {
			c.client = conn
			c.channel, err = c.client.Channel()
			if err != nil {
				log.Error().Err(err).Msg("channel cannot be connected")
				continue
			}
			log.Debug().Msg("reconnect success")
			break
		}

		log.Error().Err(err).Msg("reconnect failed")
	}
}
