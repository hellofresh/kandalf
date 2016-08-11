package workers

import (
	"sync"
	"time"

	"github.com/streadway/amqp"

	"kandalf/config"
	"kandalf/logger"
)

var (
	exchangeName string = "amq.rabbitmq.trace"
	routingKey   string = "publish.*"
)

type internalConsumer struct {
	con   *amqp.Connection
	queue *internalQueue
}

// Returns new instance of RabbitMQ consumer
func newInternalConsumer(url string, queue *internalQueue) (*internalConsumer, error) {
	con, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return &internalConsumer{
		con:   con,
		queue: queue,
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

	err = ch.ExchangeDeclare(exchangeName, "topic", true, false, true, false, nil)
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("Unable to declare exchange")

		return
	}

	q, err := ch.QueueDeclare("", false, true, false, true, nil)
	if err != nil {
		logger.Instance().
			WithError(err).
			Warning("Unable to declare queue")

		return
	}

	err = ch.QueueBind(q.Name, routingKey, exchangeName, true, nil)
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
	}

	stopConsumer := make(chan bool)

	// Now run the consumer
	go func() {
		var err error

		for m := range msgs {
			c.storeMessage(m)
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

// Build the internal message and stores it in the queue
func (c *internalConsumer) storeMessage(m amqp.Delivery) {
	msg := internalMessage{
		Body: m.Body,
	}

	if exchangeName, ok := m.Headers["exchange_name"]; ok {
		msg.ExchangeName = exchangeName.(string)
	}

	if queues, ok := m.Headers["routed_queues"]; ok {
		for _, queue := range queues.([]interface{}) {
			msg.RoutedQueues = append(msg.RoutedQueues, queue.(string))
		}
	}

	if keys, ok := m.Headers["routing_keys"]; ok {
		for _, key := range keys.([]interface{}) {
			msg.RoutingKeys = append(msg.RoutingKeys, key.(string))
		}
	}

	c.queue.add(msg)
}
