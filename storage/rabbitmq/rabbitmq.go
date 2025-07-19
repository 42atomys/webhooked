package rabbitmq

import (
	"context"
	"errors"
	"time"

	"github.com/42atomys/webhooked/internal/valuable"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type RabbitmqStorageSpec struct {
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
	MaxAttempt         int               `mapstructure:"maxAttempt" json:"maxAttempt"`

	client     *amqp.Connection
	channel    *amqp.Channel
	routingKey amqp.Queue
}

func (s *RabbitmqStorageSpec) EnsureConfigurationCompleteness() error {
	if s.DefinedContentType == "" {
		s.DefinedContentType = "text/plain"
	}

	if s.MaxAttempt == 0 {
		s.MaxAttempt = 5
	}

	return nil
}

func (s *RabbitmqStorageSpec) Initialize() error {
	var err error

	if s.client, err = amqp.Dial(s.DatabaseURL.First()); err != nil {
		return err
	}

	if s.channel, err = s.client.Channel(); err != nil {
		return err
	}

	go func() {
		for {
			reason := <-s.client.NotifyClose(make(chan *amqp.Error))
			log.Warn().Msgf("connection to rabbitmq closed, reason: %v", reason)

			s.reconnect()
		}
	}()

	if s.routingKey, err = s.channel.QueueDeclare(
		s.QueueName,
		s.Durable,
		s.DeleteWhenUnused,
		s.Exclusive,
		s.NoWait,
		nil,
	); err != nil {
		return err
	}

	return nil
}

func (s *RabbitmqStorageSpec) Store(ctx context.Context, value []byte) error {
	for attempt := 0; attempt < s.MaxAttempt; attempt++ {
		err := s.channel.Publish(
			s.Exchange,
			s.routingKey.Name,
			s.Mandatory,
			s.Immediate,
			amqp.Publishing{
				ContentType: s.DefinedContentType,
				Body:        value,
			})

		if err != nil {
			if errors.Is(err, amqp.ErrClosed) {
				log.Warn().Err(err).Msg("connection to rabbitmq closed. reconnecting...")
				s.reconnect()
				continue
			} else {
				return err
			}
		}
		return nil
	}

	return errors.New("max attempt to publish reached")

}

func (s *RabbitmqStorageSpec) reconnect() {
	for {
		// wait 1s for reconnect
		time.Sleep(time.Second)

		conn, err := amqp.Dial(s.DatabaseURL.First())
		if err == nil {
			s.client = conn
			s.channel, err = s.client.Channel()
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
