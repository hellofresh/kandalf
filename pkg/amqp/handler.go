package amqp

import (
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/hellofresh/stats-go"
	"github.com/hellofresh/stats-go/bucket"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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
func NewQueuesHandler(pipes []config.Pipe, handler MessageHandler, statsClient stats.Client) InitQueuesHandler {
	return func(conn *amqp.Connection) error {
		operation := bucket.MetricOperation{statsOpConnect, "channel"}
		channel, err := conn.Channel()
		statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
		if err != nil {
			log.WithError(err).Error("Failed to open AMQP channel")
			return err
		}

		for _, pipe := range pipes {
			operation = bucket.MetricOperation{statsOpConnect, "exchange", pipe.RabbitExchangeName}
			err = channel.ExchangeDeclare(pipe.RabbitExchangeName, exchangeTypeTopic, true, false, false, false, nil)
			statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
			if err != nil {
				log.WithError(err).Error("Failed to declare exchange")
				return err
			}

			operation = bucket.MetricOperation{statsOpConnect, "queue", pipe.RabbitQueueName}
			queue, err := channel.QueueDeclare(pipe.RabbitQueueName, pipe.RabbitDurableQueue, pipe.RabbitAutoDeleteQueue, false, true, nil)
			statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
			if err != nil {
				log.WithError(err).Error("Failed to declare queue")
				return err
			}

			operation = bucket.MetricOperation{statsOpConnect, "bind", pipe.RabbitRoutingKey}
			err = channel.QueueBind(queue.Name, pipe.RabbitRoutingKey, pipe.RabbitExchangeName, true, nil)
			statsClient.TrackOperation(statsAMQPSection, operation, nil, nil == err)
			if err != nil {
				log.WithError(err).Error("Failed to bind the queue")
				return err
			}

			operation = bucket.MetricOperation{statsOpConnect, "consume", queue.Name}
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

func consumeMessages(messages <-chan amqp.Delivery, pipe config.Pipe, handler MessageHandler, statsClient stats.Client) {
	for msg := range messages {
		err := handler(msg.Body, pipe)

		operation := bucket.MetricOperation{statsOpConsume, pipe.RabbitQueueName, pipe.RabbitRoutingKey}
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
