package workers

import (
	"path"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/Shopify/sarama.v1"

	"../config"
	"../logger"
	"../pipes"
)

type internalProducer struct {
	client    sarama.SyncProducer
	pipesList []pipes.Pipe
}

// Returns new instance of kafka producer
func newInternalProducer() (*internalProducer, error) {
	var brokers []string

	brokersInterfaces, err := config.Instance().List("kafka.brokers")
	if err != nil {
		return nil, err
	} else {
		for _, broker := range brokersInterfaces {
			brokers = append(brokers, broker.(string))
		}
	}

	cnf := sarama.NewConfig()
	cnf.Producer.RequiredAcks = sarama.WaitForAll
	cnf.Producer.Retry.Max = config.Instance().UInt("kafka.retry_max", 5)

	client, err := sarama.NewSyncProducer(brokers, cnf)
	if err != nil {
		return nil, err
	}

	return &internalProducer{
		client:    client,
		pipesList: pipes.All(),
	}, nil
}

// Sends message to the kafka
func (p *internalProducer) handleMessage(msg internalMessage) (err error) {
	var (
		msgHasExchange   bool = len(msg.Exchange) > 0
		msgHasRoutingKey bool = len(msg.RoutingKey) > 0
		pipe             pipes.Pipe
		pipeFound        bool
		exchangeFound    bool
		routingKeyFound  bool
	)

	// That's what this is all about:
	for _, p := range p.pipesList {
		pipeFound = false

		if (msgHasExchange && msgHasRoutingKey) && (p.HasExchange && p.HasRoutingKey) {
			exchangeFound, _ = path.Match(p.Exchange, msg.Exchange)
			routingKeyFound, _ = path.Match(p.RoutingKey, msg.RoutingKey)
			pipeFound = exchangeFound && routingKeyFound
		} else if (msgHasExchange && !msgHasRoutingKey) && p.HasExchange {
			pipeFound, _ = path.Match(p.Exchange, msg.Exchange)
		} else if (msgHasRoutingKey && !msgHasExchange) && p.HasRoutingKey {
			pipeFound, _ = path.Match(p.RoutingKey, msg.RoutingKey)
		}

		if pipeFound {
			pipe = p
			break
		}
	}

	if pipeFound {
		_, _, err = p.client.SendMessage(&sarama.ProducerMessage{
			Topic: pipe.Topic,
			Value: sarama.ByteEncoder(msg.Body),
		})
	} else {
		err = nil

		logger.Instance().
			WithFields(log.Fields{
				"exchange":    msg.Exchange,
				"routing_key": msg.RoutingKey,
			}).
			Debug("Unable to find Kafka topic for message")
	}

	return err
}
