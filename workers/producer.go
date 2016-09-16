package workers

import (
	"gopkg.in/Shopify/sarama.v1"

	"kandalf/config"
	"kandalf/logger"
)

type internalProducer struct {
	kafkaClient sarama.SyncProducer
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
		kafkaClient: client,
	}, nil
}

// Sends message to the kafka
func (p *internalProducer) handleMessage(msg internalMessage) (err error) {
	_, _, err = p.kafkaClient.SendMessage(&sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.ByteEncoder(msg.Body),
	})

	if err == nil {
		logger.Instance().
			WithField("topic", msg.Topic).
			Debug("Successfully sent message to kafka")
	} else {
		logger.Instance().
			WithField("topic", msg.Topic).
			Debug("An error occurred while sending message to kafka")
	}

	return err
}
