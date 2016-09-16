package workers

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"

	"kandalf/config"
	"kandalf/logger"
	"kandalf/pipes"
)

type internalConsumer struct {
	con   *amqp.Connection
	queue *internalQueue
	pipe  pipes.Pipe
}

// Returns new instance of RabbitMQ consumer
func newInternalConsumer(url string, queue *internalQueue, p pipes.Pipe) (*internalConsumer, error) {
	con, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return &internalConsumer{
		con:   con,
		queue: queue,
		pipe:  p,
	}, nil
}

// Main working cycle
func (c *internalConsumer) run(wg *sync.WaitGroup, die chan bool) {
	defer wg.Done()

	ch, err := c.con.Channel()
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("Unable to open channel")

		return
	}

	err = ch.ExchangeDeclare(c.pipe.RabbitmqExchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("Unable to declare exchange")

		return
	}

	q, err := ch.QueueDeclare(c.pipe.RabbitmqQueueName, false, true, false, true, nil)
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("Unable to declare queue")

		return
	}

	err = ch.QueueBind(q.Name, c.pipe.RabbitmqRoutingKey, c.pipe.RabbitmqExchangeName, true, nil)
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("Unable to bind the queue")

		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("Unable to start consume from the queue")

		return
	} else {
		logger.Instance().
			WithFields(log.Fields{
				"exchange_name": c.pipe.RabbitmqExchangeName,
				"queue_name":    c.pipe.RabbitmqQueueName,
				"routing_key":   c.pipe.RabbitmqRoutingKey,
			}).
			Debug("Start consuming messages")
	}

	stopConsumer := make(chan bool)

	// Now run the consumer
	go func() {
		var err error

		for m := range msgs {
			c.queue.add(internalMessage{
				Body:  m.Body,
				Topic: c.pipe.KafkaTopic,
			})
		}

		// Don't forget to close connection to RabbitMQ
		err = c.con.Close()
		if err != nil {
			logger.Instance().
				WithError(err).
				Warning("An error occurred while closing connection to RabbitMQ")
		}
	}()

	// And look into the infinite channel for interrupt message
	go func() {
		for {
			select {
			case <-die:
				stopConsumer <- true
				return
			default:
			}

			// Prevent CPU overload
			time.Sleep(config.InfiniteCycleTimeout)
		}
	}()

	<-stopConsumer

	logger.Instance().
		Debug("Will stop consuming according to the signal came from Worker")
}
