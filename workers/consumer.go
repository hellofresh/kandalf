package workers

import (
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/kandalf/pkg/config"
	"github.com/streadway/amqp"
)

type internalConsumer struct {
	con                  *amqp.Connection
	queue                *internalQueue
	pipe                 config.Pipe
	infiniteCycleTimeout time.Duration
}

// Returns new instance of RabbitMQ consumer
func newInternalConsumer(url string, queue *internalQueue, p config.Pipe, infiniteCycleTimeout time.Duration) (*internalConsumer, error) {
	con, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return &internalConsumer{
		con:                  con,
		queue:                queue,
		pipe:                 p,
		infiniteCycleTimeout: infiniteCycleTimeout,
	}, nil
}

// Main working cycle
func (c *internalConsumer) run(wg *sync.WaitGroup, die chan bool) {
	defer wg.Done()

	ch, err := c.con.Channel()
	if err != nil {
		log.WithError(err).Error("Unable to open channel")

		return
	}

	err = ch.ExchangeDeclare(c.pipe.RabbitExchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		log.WithError(err).Error("Unable to declare exchange")

		return
	}

	q, err := ch.QueueDeclare(c.pipe.RabbitQueueName, false, true, false, true, nil)
	if err != nil {
		log.WithError(err).Error("Unable to declare queue")

		return
	}

	err = ch.QueueBind(q.Name, c.pipe.RabbitRoutingKey, c.pipe.RabbitExchangeName, true, nil)
	if err != nil {
		log.WithError(err).Error("Unable to bind the queue")

		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.WithError(err).Error("Unable to start consume from the queue")

		return
	} else {
		log.
			WithFields(log.Fields{
				"exchange_name": c.pipe.RabbitExchangeName,
				"queue_name":    c.pipe.RabbitQueueName,
				"routing_key":   c.pipe.RabbitRoutingKey,
			}).
			Error("Start consuming messages")
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
			log.WithError(err).Error("An error occurred while closing connection to RabbitMQ")
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
			time.Sleep(c.infiniteCycleTimeout)
		}
	}()

	<-stopConsumer

	log.Info("Will stop consuming according to the signal came from Worker")
}
