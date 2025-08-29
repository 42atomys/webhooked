package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/42atomys/webhooked/internal/valuable"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type RabbitmqStorageSpec struct {
	DatabaseURL valuable.Valuable `mapstructure:"databaseUrl" json:"databaseUrl"`
	MaxAttempt  int               `mapstructure:"maxAttempt" json:"maxAttempt"`
	// QueueDeclare
	QueueName        string `mapstructure:"queueName" json:"queueName"`
	Durable          *bool  `mapstructure:"durable" json:"durable"`
	DeleteWhenUnused bool   `mapstructure:"deleteWhenUnused" json:"deleteWhenUnused"`
	Exclusive        bool   `mapstructure:"exclusive" json:"exclusive"`
	NoWait           bool   `mapstructure:"noWait" json:"noWait"`
	// Publish
	Exchange           string `mapstructure:"exchange" json:"exchange"`
	DefinedContentType string `mapstructure:"contentType" json:"contentType"`
	Mandatory          bool   `mapstructure:"mandatory" json:"mandatory"`
	Immediate          bool   `mapstructure:"immediate" json:"immediate"`

	client  *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func (s *RabbitmqStorageSpec) EnsureConfigurationCompleteness() error {
	if s.DefinedContentType == "" {
		s.DefinedContentType = "text/plain"
	}

	if s.MaxAttempt == 0 {
		s.MaxAttempt = 5
	}

	if s.Durable == nil {
		durable := true
		s.Durable = &durable
	}

	return nil
}

func (s *RabbitmqStorageSpec) Initialize() error {
	var err error

	if s.client, err = amqp.Dial(s.DatabaseURL.First()); err != nil {
		return fmt.Errorf("error connecting to rabbitmq: %w", err)
	}

	if s.channel, err = s.client.Channel(); err != nil {
		return fmt.Errorf("error creating channel: %w", err)
	}

	go func() {
		for {
			reason := <-s.client.NotifyClose(make(chan *amqp.Error))
			log.Warn().Msgf("connection to rabbitmq closed, reason: %v", reason)

			s.reconnect()
		}
	}()

	if s.queue, err = s.channel.QueueDeclare(
		s.QueueName,
		*s.Durable,
		s.DeleteWhenUnused,
		s.Exclusive,
		s.NoWait,
		nil,
	); err != nil {
		return fmt.Errorf("error declaring queue: %w", err)
	}

	return nil
}

func (s *RabbitmqStorageSpec) Store(ctx context.Context, value []byte) error {
	for attempt := 0; attempt < s.MaxAttempt; attempt++ {
		err := s.channel.PublishWithContext(
			ctx,
			s.Exchange,
			s.queue.Name,
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
				return fmt.Errorf("error publishing to rabbitmq: %w", err)
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
