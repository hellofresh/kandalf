package amqp

import (
	"github.com/hellofresh/stats-go/bucket"
	"github.com/hellofresh/stats-go/client"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"

	"github.com/hellofresh/kandalf/pkg/config"
)

const (
	exchangeTypeTopic = "topic"

	statsAMQPSection = "amqp"
	statsOpConnect   = "connect"
	statsOpConsume   = "consume"
)

// MessageHandler is a handler function type for consumed messages
type MessageHandler func(body []byte, pipe config.Pipe) error

// NewQueuesHandler instantiates queues initialisation handler
func NewQueuesHandler(pipes []config.Pipe, handler MessageHandler, statsClient client.Client) InitQueuesHandler {
	return func(conn *amqp.Connection) error {
		operation := bucket.NewMetricOperation(statsOpConnect, "channel")
		channel, err := conn.Channel()
		statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
		if err != nil {
			log.WithError(err).Error("Failed to open AMQP channel")
			return err
		}

		for _, pipe := range pipes {
			operation = bucket.NewMetricOperation(statsOpConnect, "exchange", pipe.RabbitExchangeName)
			err = channel.ExchangeDeclare(
				pipe.RabbitExchangeName,
				exchangeTypeTopic,
				!pipe.RabbitTransientExchange,
				false,
				false,
				false,
				nil,
			)
			statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
			if err != nil {
				log.WithError(err).Error("Failed to declare exchange")
				return err
			}

			operation = bucket.NewMetricOperation(statsOpConnect, "queue", pipe.RabbitQueueName)
			queue, err := channel.QueueDeclare(pipe.RabbitQueueName, pipe.RabbitDurableQueue, pipe.RabbitAutoDeleteQueue, false, true, nil)
			statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
			if err != nil {
				log.WithError(err).Error("Failed to declare queue")
				return err
			}

			for i := range pipe.RabbitRoutingKey {
				operation = bucket.NewMetricOperation(statsOpConnect, "bind", pipe.RabbitRoutingKey[i])
				err = channel.QueueBind(queue.Name, pipe.RabbitRoutingKey[i], pipe.RabbitExchangeName, true, nil)
				statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
				if err != nil {
					log.WithError(err).Error("Failed to bind the queue")
					return err
				}
			}

			ch, err := channel.Consume(queue.Name, queue.Name+"_consumer", false, false, false, false, nil)
			statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
			if err != nil {
				log.WithError(err).Error("Failed to register a consumer")
				return nil
			}

			go consumeMessages(ch, pipe, handler, statsClient)
		}

		return nil
	}
}

func consumeMessages(messages <-chan amqp.Delivery, pipe config.Pipe, handler MessageHandler, statsClient client.Client) {
	for msg := range messages {
		err := handler(msg.Body, pipe)

		operation := bucket.NewMetricOperation(statsOpConsume, pipe.RabbitQueueName)
		statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
		if err != nil {
			log.WithError(err).WithField("pipe", pipe.String()).
				Error("Failed to consume AMQP message")
			if err = msg.Nack(false, true); err != nil {
				log.WithError(err).WithField("pipe", pipe.String()).Error("Failed to NAck AMQP message")
			}
		} else {
			if err = msg.Ack(false); err != nil {
				log.WithError(err).WithField("pipe", pipe.String()).Error("Failed to NAck AMQP message")
			}
		}
	}
}
