package amqp

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

// InitQueuesHandler is a handler function type for AMQP connection
type InitQueuesHandler func(conn *amqp.Connection) error

// Connection struct holds data for AMQP connection
type Connection struct {
	dsn        string
	initQueues InitQueuesHandler
	conn       *amqp.Connection
}

// NewConnection instantiates and establishes new AMQP connection
func NewConnection(dsn string, initQueues InitQueuesHandler) (*Connection, error) {
	c := &Connection{dsn, initQueues, nil}

	if err := c.establishConnection(); nil != err {
		return c, err
	}
	if err := c.initQueues(c.conn); nil != err {
		return c, err
	}
	c.initNotifyClose()

	return c, nil
}

// Close closes AMQP connection
func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) establishConnection() error {
	var err error

	log.WithField("dsn", c.dsn).Info("Establishing RabbitMQ connection")
	if c.conn, err = amqp.Dial(c.dsn); nil != err {
		return err
	}

	return nil
}

func (c *Connection) initNotifyClose() {
	go func() {
		shutdownError := <-c.conn.NotifyClose(make(chan *amqp.Error))
		log.WithField("error", shutdownError).Error("Caught AMQP close notification")
		if nil != shutdownError {
			log.WithField("timeout", c.conn.Config.Heartbeat).
				Info("Caught AMQP close notification with error, trying to reconnect")
			c.reEstablishConnection(c.conn.Config.Heartbeat)
		}
	}()
}

func (c *Connection) reEstablishConnection(timeout time.Duration) {
	for {
		time.Sleep(timeout)
		if err := c.establishConnection(); err != nil {
			log.WithError(err).WithField("timeout", timeout).
				Error("Failed to establish new connection, will try later")
		} else {
			if err := c.initQueues(c.conn); nil != err {
				log.WithError(err).WithField("timeout", timeout).
					Error("Failed to init new queues, will try later")
			} else {
				c.initNotifyClose()
				break
			}
		}
	}
}
